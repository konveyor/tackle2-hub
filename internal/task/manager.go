package task

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"sync"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/auth"
	k8s2 "github.com/konveyor/tackle2-hub/internal/k8s"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/metrics"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/model/reflect"
	"github.com/konveyor/tackle2-hub/shared/settings"
	taskapi "github.com/konveyor/tackle2-hub/shared/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8r "k8s.io/apimachinery/pkg/runtime"
	k8j "k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8y "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// States
// also used as events:
// - Postponed
// - QuotaBlocked
const (
	Created      = taskapi.Created
	Ready        = taskapi.Ready
	Postponed    = taskapi.Postponed
	QuotaBlocked = taskapi.QuotaBlocked
	Pending      = taskapi.Pending
	Running      = taskapi.Running
	Succeeded    = taskapi.Succeeded
	Failed       = taskapi.Failed
	Canceled     = taskapi.Canceled
)

// Events
const (
	AddonSelected    = taskapi.AddonSelected
	ExtSelected      = taskapi.ExtSelected
	ImageError       = taskapi.ImageError
	PodNotFound      = taskapi.PodNotFound
	PodCreated       = taskapi.PodCreated
	PodPending       = taskapi.PodPending
	PodUnschedulable = taskapi.PodUnschedulable
	PodRunning       = taskapi.PodRunning
	PodSucceeded     = taskapi.PodSucceeded
	PodFailed        = taskapi.PodFailed
	PodDeleted       = taskapi.PodDeleted
	Escalated        = taskapi.Escalated
	Released         = taskapi.Released
	ContainerKilled  = taskapi.ContainerKilled
)

// Mode
const (
	Batch    = taskapi.Batch
	Pipeline = taskapi.Pipeline
)

// k8s labels.
const (
	TaskLabel = "task"
	AppLabel  = "app"
	RoleLabel = "role"
)

const (
	Unit = time.Second
)

const (
	Addon  = "addon"
	Shared = "shared"
	Cache  = "cache"
)

var (
	IsRegex = regexp.MustCompile("[^0-9A-Za-z_-]")
)

var (
	Settings = &settings.Settings
	Log      = logr.New("task-scheduler", Settings.Log.Task)
)

// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
	// Addon token scopes.
	Scopes []string
	// cluster resources.
	cluster Cluster
	// queue of actions.
	queue chan func()
	// logManager provides pod log collection.
	logManager LogManager
	// pod capacity monitor
	capacity CapacityMonitor
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	m.queue = make(chan func(), 100)
	m.cluster.Client = m.Client
	m.logManager = LogManager{
		collector: make(map[string]*LogCollector),
		DB:        m.DB,
	}
	m.capacity.Run(ctx, &m.cluster)
	auth.Validators = append(
		auth.Validators,
		&Validator{
			Client: m.Client,
		})
	if Log.V(1).Enabled() {
		m.DB = m.DB.Debug()
	}
	go func() {
		Log.Info("Manager started.")
		defer Log.Info("Done.")
		for {
			select {
			case <-ctx.Done():
				Log.Info("Manager stopped.")
				return
			default:
				err := m.cluster.Refresh()
				if err == nil {
					m.deleteOrphanPods()
					m.deleteRetainedPods()
					m.runActions()
					m.updateRunning(ctx)
					m.deleteZombies()
					m.startReady()
					m.pause()
				} else {
					if errors.Is(err, &NotReconciled{}) {
						Log.Info(err.Error())
					} else {
						Log.Error(err, "")
					}
					m.pause()
				}
			}
		}
	}()
}

// Create a task.
func (m *Manager) Create(db *gorm.DB, requested *Task) (err error) {
	err = m.findRefs(requested)
	if err != nil {
		return
	}
	task := NewTask(&model.Task{})
	switch requested.State {
	case "":
		requested.State = Created
		fallthrough
	case Created,
		Ready:
		task.CreateUser = requested.CreateUser
		task.Name = requested.Name
		task.Kind = requested.Kind
		task.Addon = requested.Addon
		task.Extensions = requested.Extensions
		task.State = requested.State
		task.Locator = requested.Locator
		task.Priority = requested.Priority
		task.Policy = requested.Policy
		task.TTL = requested.TTL
		task.Data = requested.Data
		task.ApplicationID = requested.ApplicationID
		task.PlatformID = requested.PlatformID
		task.BucketID = requested.BucketID
		task.TaskGroupID = requested.TaskGroupID
	default:
		err = &BadRequest{
			Reason: "state must be (Created|Ready)",
		}
		return
	}
	db = db.Omit(clause.Associations)
	err = db.Create(task).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	requested.Task = task.Task
	return
}

// Update update task.
func (m *Manager) Update(db *gorm.DB, requested *Task) (err error) {
	found := &Task{}
	err = db.First(found, requested.ID).Error
	if err != nil {
		return
	}
	switch found.State {
	case Created:
		db = reflect.Select(
			db,
			requested,
			"UpdateUser",
			"Name",
			"Kind",
			"Addon",
			"Extensions",
			"State",
			"Locator",
			"Priority",
			"Policy",
			"TTL",
			"Data",
			"ApplicationID",
			"PlatformID",
			"TaskGroupID")
		err = m.findRefs(requested)
		if err != nil {
			return
		}
		db = db.Where("State", Created)
		err = db.Save(requested).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case Ready,
		Pending,
		QuotaBlocked,
		Postponed:
		db = reflect.Select(
			db,
			requested,
			"UpdateUser",
			"Name",
			"Locator",
			"Policy",
			"TTL")
		db = db.Where(
			"state IN (?)",
			[]string{
				Ready,
				Pending,
				QuotaBlocked,
				Postponed,
			})
		err = db.Save(requested).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		// discarded.
		return
	}
	return
}

// Delete a task.
func (m *Manager) Delete(db *gorm.DB, id uint) (err error) {
	task := &Task{}
	err = db.First(task, id).Error
	if err != nil {
		return
	}
	m.action(
		func() (err error) {
			err = task.Delete(m.Client)
			if err != nil {
				return
			}
			err = m.DB.Delete(task).Error
			return
		})
	return
}

// Cancel a task.
func (m *Manager) Cancel(db *gorm.DB, id uint) (err error) {
	task := &Task{}
	err = db.First(task, id).Error
	if err != nil {
		return
	}
	m.action(
		func() (err error) {
			switch task.State {
			case Succeeded,
				Failed,
				Canceled:
				// discarded.
				return
			default:
			}
			pod, found := m.cluster.Pod(path.Base(task.Pod))
			if found {
				snErr := m.podSnapshot(task, pod)
				Log.Error(
					snErr,
					"Snapshot not created.")
			}
			err = task.Cancel(m.Client)
			if err != nil {
				return
			}
			err = task.update(m.DB)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			return
		})
	return
}

// Pause.
func (m *Manager) pause() {
	d := Unit * time.Duration(Settings.Frequency.Task)
	time.Sleep(d)
}

// action enqueues an asynchronous action.
func (m *Manager) action(action func() error) {
	m.queue <- func() {
		var err error
		defer func() {
			p := recover()
			if p != nil {
				if pErr, cast := p.(error); cast {
					err = pErr
				}
			}
			if err != nil {
				Log.Error(err, "Action failed.")
			}
		}()
		err = action()
	}
	return
}

// runActions executes queued actions.
func (m *Manager) runActions() {
	d := time.Millisecond * 10
	for {
		select {
		case action := <-m.queue:
			action()
		case <-time.After(d):
			return
		}
	}
}

// startReady starts ready tasks.
func (m *Manager) startReady() {
	var err error
	defer func() {
		Log.Error(err, "")
	}()
	//
	mark := time.Now()
	defer func() {
		d := time.Since(mark)
		threshold := 3 * time.Second
		if d > threshold {
			Log.Info(
				"Ready tasks started.",
				"duration",
				d.String(),
				"threshold",
				threshold.String())
		}
	}()
	if m.capacity.Exceeded() {
		Log.Info("Capacity exceeded: pod creation paused.")
		return
	}
	quota := &Quota{}
	quota.with(&m.cluster)
	fetched := []*model.Task{}
	db := m.DB.Order("priority DESC, id")
	result := db.Find(
		&fetched,
		"state IN ?",
		[]string{
			Ready,
			Postponed,
			QuotaBlocked,
			Pending,
		})
	if result.Error != nil {
		return
	}
	if len(fetched) == 0 {
		return
	}
	list := m.taskList(fetched)
	list, err = m.disabled(list)
	if err != nil {
		return
	}
	err = m.adjustPriority(list)
	if err != nil {
		return
	}
	list, err = m.selectAddons(list)
	if err != nil {
		return
	}
	err = m.postpone(list)
	if err != nil {
		return
	}
	err = m.batchUpdate(list)
	if err != nil {
		return
	}
	err = m.createPod(list, quota)
	if err != nil {
		return
	}
	err = m.batchUpdate(list)
	if err != nil {
		return
	}
	return
}

// disabled fails tasks when tasking is not enabled.
// The returned list is empty when disabled.
func (m *Manager) disabled(list []*Task) (kept []*Task, err error) {
	if Settings.Hub.Task.Enabled {
		kept = list
		return
	}
	for _, task := range list {
		mark := time.Now()
		task.State = Failed
		task.Terminated = &mark
		task.Error("Error", "Tasking is disabled.")
	}
	return
}

// FindRefs find referenced resources.
// - addon
// - extensions
// - kind
// - priority
// The priority is defaulted to the kind as needed.
func (m *Manager) findRefs(task *Task) (err error) {
	if !Settings.Hub.Task.Enabled {
		return
	}
	if task.Addon != "" {
		_, found := m.cluster.Addon(task.Addon)
		if !found {
			err = &AddonNotFound{Name: task.Addon}
			return
		}
	}
	for _, name := range task.Extensions {
		_, found := m.cluster.Extension(name)
		if !found {
			err = &ExtensionNotFound{Name: name}
			return
		}
	}
	if task.Kind == "" {
		return
	}
	kind, found := m.cluster.Task(task.Kind)
	if !found {
		err = &KindNotFound{Name: task.Kind}
		return
	}
	if task.Priority == 0 {
		task.Priority = kind.Spec.Priority
	}
	other := model.Data{Any: kind.Data()}
	merged := task.Data.Merge(other)
	if !merged {
		task.Data = other
	}
	return
}

// selectAddon selects addon as needed.
// The returned list has failed tasks removed.
func (m *Manager) selectAddons(list []*Task) (kept []*Task, err error) {
	if len(list) == 0 {
		return
	}
	mark := time.Now()
	var addon *crd.Addon
	for _, task := range list {
		addon, err = m.selectAddon(task)
		if err == nil {
			err = m.selectExtensions(task, addon)
		}
		if err != nil {
			matched, _ := SoftErr(err)
			if matched {
				task.Error("Error", err.Error())
				task.Terminated = &mark
				task.State = Failed
			}
		} else {
			kept = append(kept, task)
		}
	}
	return
}

// selectAddon select an addon when not specified.
func (m *Manager) selectAddon(task *Task) (addon *crd.Addon, err error) {
	if task.Addon != "" {
		found := false
		addon, found = m.cluster.Addon(task.Addon)
		if !found {
			err = &AddonNotFound{task.Addon}
		}
		return
	}
	kind, found := m.cluster.Task(task.Kind)
	if !found {
		err = &KindNotFound{task.Kind}
		return
	}
	matched := false
	var selected *crd.Addon
	selector := NewSelector(m.DB, task)
	for _, addon = range m.cluster.Addons() {
		matched, err = task.matchTask(addon, kind)
		if err != nil {
			return
		}
		if !matched {
			continue
		}
		matched, err = selector.Match(addon.Spec.Selector)
		if err != nil {
			return
		}
		if matched {
			selected = addon
			break
		}
	}
	if selected == nil {
		err = &AddonNotSelected{}
		return
	}
	if !selected.Ready() {
		err = &NotReady{
			Kind: "Addon",
			Name: selected.Name,
		}
		return
	}
	task.Addon = selected.Name
	task.Event(AddonSelected, selected)
	return
}

// selectExtensions select extensions when not specified.
func (m *Manager) selectExtensions(task *Task, addon *crd.Addon) (err error) {
	if len(task.Extensions) > 0 {
		return
	}
	matched := false
	selector := NewSelector(m.DB, task)
	for _, extension := range m.cluster.Extensions() {
		matched, err = task.matchAddon(extension, addon)
		if err != nil {
			return
		}
		if !matched {
			continue
		}
		matched, err = selector.Match(extension.Spec.Selector)
		if err != nil {
			return
		}
		if matched {
			task.Extensions = append(task.Extensions, extension.Name)
			task.Event(ExtSelected, extension.Name)
		}
	}
	return
}

// postpone Postpones a task as needed based on rules.
// postpone order:
// - priority (lower)
// - Age (newer)
func (m *Manager) postpone(list []*Task) (err error) {
	if len(list) == 0 {
		return
	}
	sort.Slice(
		list,
		func(i, j int) bool {
			it := list[i]
			jt := list[j]
			return it.Priority < jt.Priority ||
				(it.Priority == jt.Priority &&
					it.ID > jt.ID)
		})
	postponed := map[uint]any{}
	released := map[uint]any{}
	ruleSet := []Rule{
		&RuleIsolated{},
		&RuleUnique{
			matched: make(map[uint]uint),
		},
		&RuleDeps{
			cluster: &m.cluster,
		},
	}
	domain := NewDomain(list)
	for _, task := range list {
		if !task.StateIn(Ready, Postponed, QuotaBlocked) {
			continue
		}
		var matched bool
		var reason string
		for _, rule := range ruleSet {
			matched, reason = rule.Match(task, domain)
			if matched {
				postponed[task.ID] = reason
				break
			}
		}
		if !matched && task.State == Postponed {
			released[task.ID] = 0
		}
	}
	if len(postponed)+len(released) == 0 {
		return
	}
	for _, task := range list {
		reason, found := postponed[task.ID]
		if found {
			task.State = Postponed
			task.Event(Postponed, reason)
			Log.Info(
				"Task postponed.",
				"id",
				task.ID,
				"reason",
				reason)
		}
		_, found = released[task.ID]
		if found {
			task.State = Ready
		}
	}
	return
}

// adjustPriority escalate as needed.
// To prevent priority inversion, the priority of a task's
// dependencies will be escalated provided the dependency has:
// - state of: (Ready|Pending|Postponed|QuotaBlocked)
// - priority (lower).
// When adjusted, Pending tasks pods deleted and made Ready again.
func (m *Manager) adjustPriority(list []*Task) (err error) {
	if len(list) == 0 {
		return
	}
	pE := Priority{cluster: &m.cluster}
	escalated := pE.Escalate(list)
	for _, task := range escalated {
		if task.State != Pending {
			continue
		}
		err = task.Delete(m.Client)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		task.State = Ready
	}
	return
}

// createPod creates a pod for the task.
func (m *Manager) createPod(list []*Task, quota *Quota) (err error) {
	if len(list) == 0 {
		return
	}
	sort.Slice(
		list,
		func(i, j int) bool {
			it := list[i]
			jt := list[j]
			return it.Priority > jt.Priority ||
				(it.Priority == jt.Priority &&
					it.ID < jt.ID)
		})
	created := 0
	capacity := max(m.capacity.Current(), 1)
	scheduled := len(m.cluster.TaskPodsScheduled())
	requested := len(list)
	for _, task := range list {
		if !task.StateIn(Ready, QuotaBlocked) {
			requested--
			continue
		}
		if scheduled >= capacity {
			break
		}
		ready := task
		started := false
		started, err = ready.Run(&m.cluster, quota)
		if err != nil {
			Log.Error(err, "")
			return
		}
		if started {
			scheduled++
			created++
			Log.Info("Task started.", "id", ready.ID)
			if ready.Retries == 0 {
				metrics.TasksInitiated.Inc()
			}
		} else {
			break
		}
	}
	Log.Info(
		"Task pods created.",
		"requested",
		requested,
		"scheduled",
		scheduled,
		"capacity",
		capacity,
		"quota",
		quota.string(),
		"created",
		created)
	return
}

// updateRunning tasks to reflect pod state.
func (m *Manager) updateRunning(ctx context.Context) {
	var err error
	defer func() {
		Log.Error(err, "")
	}()
	//
	mark := time.Now()
	defer func() {
		d := time.Since(mark)
		threshold := 3 * time.Second
		if d > threshold {
			Log.Info(
				"Running tasks updated.",
				"duration",
				d.String(),
				"threshold",
				threshold.String())
		}
	}()
	//
	fetched := []*model.Task{}
	db := m.DB.Order("priority DESC, id")
	result := db.Find(
		&fetched,
		"state IN ?",
		[]string{
			Pending,
			Running,
		})
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	if len(fetched) == 0 {
		return
	}
	var updated []*Task
	for _, task := range m.taskList(fetched) {
		pod, found := task.Reflect(&m.cluster)
		if found {
			err = m.logManager.EnsureCollection(task, pod, ctx)
			if err != nil {
				Log.Error(err, "")
				continue
			}
			if task.StateIn(Succeeded, Failed) {
				Log.Info(
					"Task completed.",
					"id",
					task.ID)
				err = m.podSnapshot(task, pod)
				if err != nil {
					Log.Error(err, "")
					continue
				}
				podRetention := task.podRetention()
				if podRetention > 0 {
					err = m.ensureTerminated(task, pod)
					if err != nil {
						podRetention = 0
					}
				}
				if podRetention == 0 {
					err = task.Delete(m.Client)
					if err != nil {
						Log.Error(err, "")
						continue
					}
				} else {
					task.Retained = true
				}
				var advanced []*Task
				advanced, err = m.advancePipeline(task)
				if err != nil {
					Log.Error(err, "")
					continue
				}
				updated = append(updated, advanced...)
			}
		}
		updated = append(updated, task)
	}
	err = m.batchUpdate(updated)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
}

// deleteRetained deletes expired retained tasks.
func (m *Manager) deleteRetainedPods() {
	var err error
	defer func() {
		Log.Error(err, "")
	}()
	fetched := []*model.Task{}
	err = m.DB.Find(
		&fetched,
		"Retained",
		true).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	var updated []*Task
	for _, task := range m.taskList(fetched) {
		if !task.podRetentionExpired() {
			continue
		}
		err = task.Delete(m.Client)
		if err != nil {
			Log.Error(err, "")
			continue
		}
		task.Retained = false
		updated = append(updated, task)
	}
	err = m.batchUpdate(updated)
	if err != nil {
		Log.Error(err, "")
		return
	}
}

// deleteZombies - detect and delete zombie pods.
// A zombie is a (succeed|failed) task with a running pod that
// the manager has previously tried to kill.
func (m *Manager) deleteZombies() {
	var err error
	defer func() {
		Log.Error(err, "")
	}()
	var pods []string
	for _, pod := range m.cluster.TaskPods() {
		if pod.Status.Phase == core.PodRunning {
			ref := path.Join(pod.Namespace, pod.Name)
			pods = append(
				pods,
				ref)
		}
	}
	fetched := []*Task{}
	db := m.DB.Select("Events")
	err = db.Find(
		&fetched,
		"state IN ? and pod IN ?",
		[]string{
			Succeeded,
			Failed,
		},
		pods).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, task := range fetched {
		event, found := task.LastEvent(ContainerKilled)
		if !found {
			continue
		}
		if time.Since(event.Last) > time.Minute {
			Log.Info(
				"Zombie detected.",
				"task",
				task.ID,
				"pod",
				task.Pod)
			err = task.Delete(m.Client)
			if err != nil {
				Log.Error(err, "")
			}
		}
	}
}

// deleteOrphanPods finds and deletes task pods not referenced by a task.
func (m *Manager) deleteOrphanPods() {
	var err error
	defer func() {
		Log.Error(err, "")
	}()
	owned := make(map[string]byte)
	list := []*Task{}
	db := m.DB.Select("pod")
	err = db.Find(&list, "pod != ''").Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, task := range list {
		owned[task.Pod] = 0
	}
	for _, pod := range m.cluster.TaskPods() {
		ref := path.Join(pod.Namespace, pod.Name)
		if _, found := owned[ref]; !found {
			Log.Info("Orphan pod found.", "ref", ref)
			task := NewTask(&model.Task{})
			task.Pod = ref
			err = task.Delete(m.Client)
			if err != nil {
				err = liberr.Wrap(err)
				Log.Error(err, "")
			}
		}
	}
}

// podSnapshot attaches a pod description and logs.
// Includes:
//   - pod YAML
//   - pod Events
func (m *Manager) podSnapshot(task *Task, pod *core.Pod) (err error) {
	events, err := m.podEvents(pod)
	if err != nil {
		return
	}
	file := &model.File{Name: "pod.yaml"}
	err = m.DB.Create(file).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	f, err := os.Create(file.Path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	serializer := k8j.NewSerializerWithOptions(
		k8y.DefaultMetaFactory,
		nil,
		nil,
		k8j.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: false,
		})
	pod.ManagedFields = nil
	format := "  %-8s%-11s%-6s%-19s%s\n"
	_, _ = f.WriteString("---\n")
	b, _ := k8r.Encode(serializer, pod)
	_, _ = f.Write(b)
	_, _ = f.WriteString("\n---\n")
	_, _ = f.WriteString("Events: |\n")
	_, _ = f.WriteString(
		fmt.Sprintf(
			format,
			"Type",
			"Reason",
			"Age",
			"Reporter",
			"Message"))
	_, _ = f.WriteString(
		fmt.Sprintf(
			format,
			"-------",
			"----------",
			"-----",
			"------------------",
			"------------------"))
	for _, event := range events {
		_, _ = f.WriteString(
			fmt.Sprintf(
				format,
				event.Type,
				event.Reason,
				event.Age,
				event.Reporter,
				event.Message))
	}
	task.attach(file)
	return
}

// podEvents get pod events.
func (m *Manager) podEvents(pod *core.Pod) (events []Event, err error) {
	clientSet, err := k8s2.NewClientSet()
	if err != nil {
		return
	}
	options := meta.ListOptions{
		FieldSelector: "involvedObject.name=" + pod.Name,
		TypeMeta: meta.TypeMeta{
			Kind: "Pod",
		},
	}
	eventClient := clientSet.CoreV1().Events(Settings.Hub.Namespace)
	eventList, err := eventClient.List(context.TODO(), options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, event := range eventList.Items {
		duration := event.LastTimestamp.Sub(event.FirstTimestamp.Time)
		events = append(
			events,
			Event{
				Type:     event.Type,
				Reason:   event.Reason,
				Age:      duration.String(),
				Reporter: event.ReportingController,
				Message:  event.Message,
			})
	}
	return
}

// ensureTerminated - Terminate running containers.
func (m *Manager) ensureTerminated(task *Task, pod *core.Pod) (err error) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil {
			continue
		}
		if status.Started == nil || !*status.Started {
			continue
		}
		err = m.terminateContainer(task, pod, status.Name)
		if err != nil {
			return
		}
	}
	return
}

// terminateContainer - Terminate container as needed.
// The container is killed.
// Should the container continue to run after (1) minute,
// it is reported as an error.
func (m *Manager) terminateContainer(task *Task, pod *core.Pod, container string) (err error) {
	Log.V(1).Info("KILL container", "container", container)
	clientSet, err := k8s2.NewClientSet()
	if err != nil {
		return
	}
	cmd := []string{
		"sh",
		"-c",
		"kill 1",
	}
	req := clientSet.CoreV1().RESTClient().Post()
	req = req.Resource("pods")
	req = req.Name(pod.Name)
	req = req.Namespace(pod.Namespace)
	req = req.SubResource("exec")
	option := &core.PodExecOptions{
		Command:   cmd,
		Container: container,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	cfg, _ := config.GetConfig()
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return
	}
	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		Log.Info(
			"Container KILL failed.",
			"name",
			container,
			"err",
			err.Error(),
			"stderr",
			stderr.String())
	} else {
		task.Event(
			ContainerKilled,
			"container: '%s' had not terminated.",
			container)
		Log.Info(
			"Container KILLED.",
			"pod",
			pod.Name,
			"name",
			container)
	}
	return
}

// advancePipeline makes the next task in a mode=pipeline task group Ready.
func (m *Manager) advancePipeline(task *Task) (updated []*Task, err error) {
	if task.TaskGroupID == nil {
		return
	}
	pipelineSet := PipelineSet{}
	err = pipelineSet.Load(m.DB)
	if err != nil {
		return
	}
	if !pipelineSet.Contains(task) {
		return
	}
	var tasks []*Task
	db := m.DB.Order("ID")
	err = db.Find(&tasks, "TaskGroupID", task.TaskGroupID).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, member := range tasks {
		if task.ID == member.ID {
			continue
		}
		switch task.State {
		case Succeeded:
			switch member.State {
			case "", Created:
				member.State = Ready
				updated = append(updated, member)
				return
			default:
				// next
			}
		case Failed:
			switch member.State {
			case Succeeded,
				Failed,
				Canceled:
			default:
				reason := fmt.Sprintf(
					"Canceled:%d, when (pipelined) task:%d failed.",
					member.ID,
					task.ID)
				member.Event(Canceled, reason)
				nErr := member.Cancel(m.Client)
				if nErr != nil {
					nErr = liberr.Wrap(nErr)
					Log.Error(nErr, "")
				}
				updated = append(updated, member)
			}
		default:
			return
		}
	}
	return
}

// batchUpdate tasks.
func (m *Manager) batchUpdate(tasks []*Task) (err error) {
	err = m.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			for _, task := range tasks {
				if task.hasChanged() {
					err = task.update(tx)
					if err != nil {
						err = liberr.Wrap(err)
						break
					}
				}
			}
			return
		})
	if err == nil {
		return
	}
	for _, task := range tasks {
		if task.hasChanged() {
			err = task.update(m.DB)
			if err != nil {
				err = liberr.Wrap(err)
				break
			}
		}
	}
	return
}

// taskList returns a list of Task.
func (m *Manager) taskList(in []*model.Task) (out []*Task) {
	out = make([]*Task, len(in))
	for i := range in {
		out[i] = NewTask(in[i])
	}
	return
}

// Priority escalator.
type Priority struct {
	cluster *Cluster
}

// Escalate task dependencies as needed.
func (p *Priority) Escalate(ready []*Task) (escalated []*Task) {
	sort.Slice(
		ready,
		func(i, j int) bool {
			it := ready[i]
			jt := ready[j]
			return it.Priority > jt.Priority
		})
	for _, task := range ready {
		dependencies := p.graph(task, ready)
		for _, d := range dependencies {
			if !d.StateIn(
				Ready,
				Pending,
				Postponed,
				QuotaBlocked) {
				continue
			}
			if d.Priority < task.Priority {
				d.Priority = task.Priority
				reason := fmt.Sprintf(
					"Escalated:%d, by:%d",
					d.ID,
					task.ID)
				d.Event(Escalated, reason)
				Log.Info(reason)
				escalated = append(
					escalated,
					d)
			}
		}
	}
	escalated = p.unique(escalated)
	return
}

// graph builds a dependency graph.
func (p *Priority) graph(task *Task, ready []*Task) (deps []*Task) {
	kind, found := p.cluster.Task(task.Kind)
	if !found {
		return
	}
	for _, d := range kind.Spec.Dependencies {
		for _, r := range ready {
			if r.ID == task.ID {
				continue
			}
			if r.Kind != d {
				continue
			}
			if !r.MatchSubject(task) {
				continue
			}
			deps = append(deps, r)
			deps = append(deps, p.graph(r, ready)...)
		}
	}
	return
}

// unique returns a unique list of tasks.
func (p *Priority) unique(in []*Task) (out []*Task) {
	mp := make(map[uint]*Task)
	for _, ptr := range in {
		mp[ptr.ID] = ptr
	}
	for _, ptr := range mp {
		out = append(out, ptr)
	}
	return
}

// Cluster provides cached cluster resources.
// Maps must NOT be accessed directly.
type Cluster struct {
	k8s.Client
	mutex      sync.RWMutex
	tackle     *crd.Tackle
	addons     map[string]*crd.Addon
	extensions map[string]*crd.Extension
	tasks      map[string]*crd.Task
	quotas     map[string]*core.ResourceQuota
	pods       struct {
		other map[string]*core.Pod
		tasks map[string]*core.Pod
	}
}

// Refresh the cache.
func (k *Cluster) Refresh() (err error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	if !Settings.Hub.Task.Enabled {
		k.tackle = &crd.Tackle{}
		k.addons = make(map[string]*crd.Addon)
		k.extensions = make(map[string]*crd.Extension)
		k.tasks = make(map[string]*crd.Task)
		k.pods.other = make(map[string]*core.Pod)
		k.pods.tasks = make(map[string]*core.Pod)
		k.quotas = make(map[string]*core.ResourceQuota)
		return
	}
	err = k.getTackle()
	if err != nil {
		return
	}
	err = k.getAddons()
	if err != nil {
		return
	}
	err = k.getExtensions()
	if err != nil {
		return
	}
	err = k.getTasks()
	if err != nil {
		return
	}
	err = k.getPods()
	if err != nil {
		return
	}
	err = k.getQuotas()
	if err != nil {
		return
	}
	return
}

// Tackle returns the tackle resource.
func (k *Cluster) Tackle() (r *crd.Tackle) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r = k.tackle
	return
}

// Addon returns an addon my name.
func (k *Cluster) Addon(name string) (r *crd.Addon, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.addons[name]
	return
}

// Addons returns an addon my name.
func (k *Cluster) Addons() (list []*crd.Addon) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.addons {
		list = append(list, r)
	}
	return
}

// Extension returns an extension by name.
func (k *Cluster) Extension(name string) (r *crd.Extension, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.extensions[name]
	return
}

// Extensions returns an extension my name.
func (k *Cluster) Extensions() (list []*crd.Extension) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.extensions {
		list = append(list, r)
	}
	return
}

// FindExtensions returns extensions by name.
func (k *Cluster) FindExtensions(names []string) (list []*crd.Extension, err error) {
	for _, name := range names {
		r, found := k.extensions[name]
		if !found {
			err = &ExtensionNotFound{name}
			return
		}
		list = append(list, r)
	}
	return
}

// Task returns a task by name.
func (k *Cluster) Task(name string) (r *crd.Task, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.tasks[name]
	return
}

// Pod returns a pod by name.
func (k *Cluster) Pod(name string) (r *core.Pod, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.pods.tasks[name]
	return
}

// OtherPods returns a list of non-task pods.
func (k *Cluster) OtherPods() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods.other {
		list = append(list, r)
	}
	return
}

// TaskPods returns a list of task pods.
func (k *Cluster) TaskPods() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods.tasks {
		list = append(list, r)
	}
	return
}

// TaskPodsScheduled returns a list of task pods scheduled.
func (k *Cluster) TaskPodsScheduled() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods.tasks {
		if r.Status.Phase == core.PodFailed || r.Status.Phase == core.PodSucceeded {
			continue
		}
		list = append(list, r)
	}
	return
}

// Quotas returns quotas.
func (k *Cluster) Quotas() (list []*core.ResourceQuota) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.quotas {
		list = append(list, r)
	}
	return
}

// PodQuota returns the most restricted pod quota.
func (k *Cluster) PodQuota() (quota int, found bool) {
	for _, r := range k.Quotas() {
		var qty resource.Quantity
		qty, found = r.Spec.Hard[core.ResourcePods]
		if found {
			n := int(qty.Value())
			if quota == 0 || quota > n {
				quota = n
			}
		}
	}
	return
}

// getTackle
func (k *Cluster) getTackle() (err error) {
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.TackleList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.tackle = r
		return
	}
	err = liberr.New("Tackle CR not found.")
	return
}

// getAddons
func (k *Cluster) getAddons() (err error) {
	k.addons = make(map[string]*crd.Addon)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.AddonList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.addons[r.Name] = r
		if !r.Reconciled() {
			err = &NotReconciled{
				Kind: r.Kind,
				Name: r.Name,
			}
			return
		}
	}
	return
}

// getExtensions
func (k *Cluster) getExtensions() (err error) {
	k.extensions = make(map[string]*crd.Extension)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.ExtensionList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.extensions[r.Name] = r
	}
	return
}

// getTasks kinds.
func (k *Cluster) getTasks() (err error) {
	k.tasks = make(map[string]*crd.Task)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.TaskList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.tasks[r.Name] = r
	}
	return
}

// getPods
func (k *Cluster) getPods() (err error) {
	k.pods.other = make(map[string]*core.Pod)
	k.pods.tasks = make(map[string]*core.Pod)
	selector := labels.NewSelector()
	options := &k8s.ListOptions{
		Namespace:     Settings.Namespace,
		LabelSelector: selector,
	}
	list := core.PodList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		if _, found := r.Labels[TaskLabel]; found {
			k.pods.tasks[r.Name] = r
		} else {
			k.pods.other[r.Name] = r
		}

	}
	return
}

// getQuotas
func (k *Cluster) getQuotas() (err error) {
	k.quotas = make(map[string]*core.ResourceQuota)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := core.ResourceQuotaList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.quotas[r.Name] = r
	}
	return
}

// ExtEnv returns an environment variable named namespaced to an extension.
// Format: _EXT_<extension_<var>.
func ExtEnv(extension string, envar string) (s string) {
	return taskapi.ExtEnv(extension, envar)
}

// PipelineSet is a set of TaskGroup ids for ALL groups of mode=Pipeline.
type PipelineSet map[uint]byte

// Load the set.
func (m *PipelineSet) Load(db *gorm.DB) (err error) {
	var list []*model.TaskGroup
	db = db.Select("ID", "Mode")
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, p := range list {
		if p.Mode == Pipeline {
			(*m)[p.ID] = 0
		}
	}
	return
}

// Contains returns true when set contains the group referenced by the task.
func (m *PipelineSet) Contains(task *Task) (found bool) {
	if task.TaskGroupID != nil {
		_, found = (*m)[*task.TaskGroupID]
	}
	return
}
