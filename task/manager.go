package task

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/auth"
	k8s2 "github.com/konveyor/tackle2-hub/k8s"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/metrics"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/reflect"
	"github.com/konveyor/tackle2-hub/settings"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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
	Created      = "Created"
	Ready        = "Ready"
	Postponed    = "Postponed"
	QuotaBlocked = "QuotaBlocked"
	Pending      = "Pending"
	Running      = "Running"
	Succeeded    = "Succeeded"
	Failed       = "Failed"
	Canceled     = "Canceled"
)

// Events
const (
	AddonSelected   = "AddonSelected"
	ExtSelected     = "ExtensionSelected"
	ImageError      = "ImageError"
	PodNotFound     = "PodNotFound"
	PodCreated      = "PodCreated"
	PodPending      = "PodPending"
	PodRunning      = "PodRunning"
	Preempted       = "Preempted"
	PodSucceeded    = "PodSucceeded"
	PodFailed       = "PodFailed"
	PodDeleted      = "PodDeleted"
	Escalated       = "Escalated"
	Released        = "Released"
	ContainerKilled = "ContainerKilled"
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
	Log      = logr.WithName("task-scheduler")
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
	// collector registry.
	collector map[string]*LogCollector
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	m.queue = make(chan func(), 100)
	m.collector = make(map[string]*LogCollector)
	m.cluster.Client = m.Client
	auth.Validators = append(
		auth.Validators,
		&Validator{
			Client: m.Client,
		})
	go func() {
		Log.Info("Started.")
		defer Log.Info("Done.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := m.cluster.Refresh()
				if err == nil {
					m.deleteOrphanCollector()
					m.deleteOrphanPods()
					m.runActions()
					m.updateRunning()
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
	task := &Task{&model.Task{}}
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
		task.BucketID = requested.BucketID
	default:
		err = &BadRequest{
			Reason: "state must be (Created|Ready)",
		}
		return
	}
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
			"ApplicationID")
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
			Running,
		})
	if result.Error != nil {
		return
	}
	if len(fetched) == 0 {
		return
	}
	var list []*Task
	for _, task := range fetched {
		list = append(list, &Task{task})
	}
	list, err = m.disconnected(list)
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
	err = m.createPod(list)
	if err != nil {
		return
	}
	err = m.preempt(list)
	if err != nil {
		return
	}
	return
}

// disconnected fails tasks when hub is disconnected.
// The returned list is empty when disconnected.
func (m *Manager) disconnected(list []*Task) (kept []*Task, err error) {
	if !Settings.Disconnected {
		kept = list
		return
	}
	for _, task := range list {
		mark := time.Now()
		task.State = Failed
		task.Terminated = &mark
		task.Error("Error", "Hub is disconnected.")
		err = task.update(m.DB)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
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
	if Settings.Disconnected {
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
				err = task.update(m.DB)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				err = nil
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
		if addon.Spec.Task != kind.Name {
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
		&RulePreempted{},
		&RuleUnique{
			matched: make(map[uint]uint),
		},
		&RuleDeps{
			cluster: &m.cluster,
		},
	}
	for _, task := range list {
		if !task.StateIn(Ready, Postponed, QuotaBlocked) {
			continue
		}
		ready := task
		for _, other := range list {
			if ready.ID == other.ID {
				continue
			}
			for _, rule := range ruleSet {
				matched, reason := rule.Match(ready, other)
				if matched {
					postponed[task.ID] = reason
					continue
				}
			}
		}
		_, found := postponed[task.ID]
		if !found {
			if task.State == Postponed {
				released[task.ID] = 0
			}
		}
	}
	if len(postponed)+len(released) == 0 {
		return
	}
	for _, task := range list {
		updated := false
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
			updated = true
		}
		_, found = released[task.ID]
		if found {
			task.State = Ready
			updated = true
		}
		if updated {
			err = task.update(m.DB)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
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
		err = task.update(m.DB)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// createPod creates a pod for the task.
func (m *Manager) createPod(list []*Task) (err error) {
	sort.Slice(
		list,
		func(i, j int) bool {
			it := list[i]
			jt := list[j]
			return it.Priority > jt.Priority ||
				(it.Priority == jt.Priority &&
					it.ID < jt.ID)
		})
	for _, task := range list {
		if !task.StateIn(Ready, QuotaBlocked) {
			continue
		}
		ready := task
		started := false
		started, err = ready.Run(&m.cluster)
		if err != nil {
			Log.Error(err, "")
			return
		}
		err = ready.update(m.DB)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if started {
			Log.Info("Task started.", "id", ready.ID)
			if ready.Retries == 0 {
				metrics.TasksInitiated.Inc()
			}
		}
	}
	return
}

// preempt reschedules a Running task as needed.
// The `preempted` task must be:
// - state=Running.
// - lower priority.
// The `blocked` task must be:
// - higher priority
// - pod blocked by quota or pending for a defined period.
// Preempt order:
// - priority (lowest).
// - age (newest).
// Preempt limit: 10% each pass.
func (m *Manager) preempt(list []*Task) (err error) {
	preemption := Settings.Hub.Task.Preemption
	if len(list) == 0 {
		return
	}
	mark := time.Now()
	blocked := []*Task{}
	running := []*Task{}
	preempt := []Preempt{}
	sort.Slice(
		list,
		func(i, j int) bool {
			it := list[i]
			jt := list[j]
			return it.Priority > jt.Priority ||
				(it.Priority == jt.Priority &&
					it.ID < jt.ID)
		})
	for _, task := range list {
		switch task.State {
		case Ready:
		case QuotaBlocked:
			enabled := preemption.Enabled || task.Policy.PreemptEnabled
			if !enabled {
				break
			}
			event, found := task.LastEvent(QuotaBlocked)
			if found {
				count := preemption.Delayed / time.Second
				if event.Count > int(count) {
					blocked = append(blocked, task)
				}
			}
		case Pending:
			enabled := preemption.Enabled || task.Policy.PreemptEnabled
			if !enabled {
				break
			}
			event, found := task.LastEvent(PodCreated)
			if found {
				if mark.Sub(event.Last) > preemption.Delayed {
					blocked = append(blocked, task)
				}
			}
		case Running:
			exempt := task.Policy.PreemptExempt
			if !exempt {
				running = append(running, task)
			}
		}
	}
	if len(blocked) == 0 {
		return
	}
	for _, b := range blocked {
		for _, p := range running {
			if b.Priority > p.Priority {
				preempt = append(
					preempt,
					Preempt{
						task: p,
						by:   b,
					})
			}
		}
	}
	sort.Slice(
		preempt,
		func(i, j int) bool {
			it := list[i]
			jt := list[j]
			return it.Priority < jt.Priority ||
				(it.Priority == jt.Priority &&
					it.ID > jt.ID)
		})
	n := 0
	for _, request := range preempt {
		p := request.task
		by := request.by
		reason := fmt.Sprintf(
			"Preempted:%d, by: %d",
			p.ID,
			by.ID)
		_ = p.Delete(m.Client)
		p.Pod = ""
		p.State = Ready
		p.Started = nil
		p.Terminated = nil
		p.Errors = nil
		p.Event(Preempted, reason)
		Log.Info(reason)
		err = p.update(m.DB)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		n++
		// preempt x%.
		if len(blocked)/n*100 > preemption.Rate {
			break
		}
	}
	return
}

// updateRunning tasks to reflect pod state.
func (m *Manager) updateRunning() {
	var err error
	defer func() {
		Log.Error(err, "")
	}()
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
	var list []*Task
	for _, task := range fetched {
		list = append(list, &Task{task})
	}
	for _, task := range list {
		running := task
		pod, found := running.Reflect(&m.cluster)
		if found {
			err = m.ensureCollector(task, pod)
			if err != nil {
				Log.Error(err, "")
				continue
			}
			if task.StateIn(Succeeded, Failed) {
				err = m.podSnapshot(running, pod)
				if err != nil {
					Log.Error(err, "")
					continue
				}
				podRetention := 0
				if running.State == Succeeded {
					podRetention = Settings.Hub.Task.Pod.Retention.Succeeded
				} else {
					podRetention = Settings.Hub.Task.Pod.Retention.Failed
				}
				if podRetention > 0 {
					err = m.ensureTerminated(running, pod)
					if err != nil {
						podRetention = 0
					}
				}
				if podRetention == 0 {
					err = running.Delete(m.Client)
					if err != nil {
						Log.Error(err, "")
						continue
					}
				}
			}
		}
		err = running.update(m.DB)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		Log.V(1).Info("Task updated.", "id", running.ID)
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
	for _, pod := range m.cluster.Pods() {
		if pod.Status.Phase == core.PodRunning {
			ref := path.Join(pod.Namespace, pod.Name)
			pods = append(
				pods,
				ref)
		}
	}
	fetched := []*Task{}
	db := m.DB.Select("Events")
	db = db.Where("Pod", pods)
	db = db.Where("state IN ?",
		[]string{
			Succeeded,
			Failed,
		})
	err = db.Find(&fetched).Error
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
	for _, pod := range m.cluster.Pods() {
		ref := path.Join(pod.Namespace, pod.Name)
		if _, found := owned[ref]; !found {
			Log.Info("Orphan pod found.", "ref", ref)
			task := Task{&model.Task{}}
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
	events, err := m.podEvent(pod)
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
	type Pod struct {
		core.Pod `yaml:",inline"`
		Events   []Event `yaml:",omitempty"`
	}
	d := Pod{
		Pod:    *pod,
		Events: events,
	}
	b, _ := yaml.Marshal(d)
	_, _ = f.Write(b)
	task.attach(file)
	return
}

// podEvent get pod events.
func (m *Manager) podEvent(pod *core.Pod) (events []Event, err error) {
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

// ensureCollector - ensure each container has a log collector attached.
func (m *Manager) ensureCollector(task *Task, pod *core.Pod) (err error) {
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil {
			continue
		}
		key := pod.Name + "." + container.Name
		if _, found := m.collector[key]; found {
			continue
		}
		collector := &LogCollector{
			Registry:  m.collector,
			DB:        m.DB,
			Pod:       pod,
			Container: &container,
		}
		err = collector.Begin(task)
		if err != nil {
			return
		}
		m.collector[key] = collector
	}
	return
}

// deleteOrphanCollector delete orphaned collectors.
func (m *Manager) deleteOrphanCollector() {
	for key, collector := range m.collector {
		_, found := m.cluster.Pod(collector.Pod.Name)
		if !found {
			delete(m.collector, key)
		}
	}
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
			"name",
			container)
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
			if r.ApplicationID == nil || task.ApplicationID == nil {
				continue
			}
			if *r.ApplicationID != *task.ApplicationID {
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
	pods       map[string]*core.Pod
}

// Refresh the cache.
func (k *Cluster) Refresh() (err error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	if Settings.Hub.Disconnected {
		k.tackle = &crd.Tackle{}
		k.addons = make(map[string]*crd.Addon)
		k.extensions = make(map[string]*crd.Extension)
		k.tasks = make(map[string]*crd.Task)
		k.pods = make(map[string]*core.Pod)
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

// Extensions returns an addon my name.
func (k *Cluster) Extensions() (list []*crd.Extension) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.extensions {
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
	r, found = k.pods[name]
	return
}

// Pods returns a list of pods.
func (k *Cluster) Pods() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods {
		list = append(list, r)
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
	k.pods = make(map[string]*core.Pod)
	selector := labels.NewSelector()
	req, _ := labels.NewRequirement(TaskLabel, selection.Exists, []string{})
	selector = selector.Add(*req)
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
		k.pods[r.Name] = r
	}
	return
}

// ExtEnv returns an environment variable named namespaced to an extension.
// Format: _EXT_<extension_<var>.
func ExtEnv(extension string, envar string) (s string) {
	s = strings.Join(
		[]string{
			"_EXT",
			strings.ToUpper(extension),
			envar,
		},
		"_")
	return
}

// Preempt request.
type Preempt struct {
	task *Task
	by   *Task
}

// LogCollector collect and report container logs.
type LogCollector struct {
	nBuf      int
	Registry  map[string]*LogCollector
	DB        *gorm.DB
	Pod       *core.Pod
	Container *core.ContainerStatus
	//
	nSkip int64
}

// Begin - get container log and store in file.
// - Request logs.
// - Create file resource and attach to the task.
// - Register collector.
// - Write (copy) log.
// - Unregister collector.
func (r *LogCollector) Begin(task *Task) (err error) {
	reader, err := r.request()
	if err != nil {
		return
	}
	f, err := r.file(task)
	if err != nil {
		return
	}
	go func() {
		defer func() {
			_ = reader.Close()
			_ = f.Close()
		}()
		err := r.copy(reader, f)
		Log.Error(err, "")
	}()
	return
}

// request
func (r *LogCollector) request() (reader io.ReadCloser, err error) {
	options := &core.PodLogOptions{
		Container: r.Container.Name,
		Follow:    true,
	}
	clientSet, err := k8s2.NewClientSet()
	if err != nil {
		return
	}
	podClient := clientSet.CoreV1().Pods(Settings.Hub.Namespace)
	req := podClient.GetLogs(r.Pod.Name, options)
	reader, err = req.Stream(context.TODO())
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// name returns the canonical name for the container log.
func (r *LogCollector) name() (s string) {
	s = r.Container.Name + ".log"
	return
}

// file returns an attached log file for writing.
func (r *LogCollector) file(task *Task) (f *os.File, err error) {
	f, found, err := r.find(task)
	if found || err != nil {
		return
	}
	f, err = r.create(task)
	return
}

// find finds and opens an attached log file by name.
func (r *LogCollector) find(task *Task) (f *os.File, found bool, err error) {
	var file model.File
	name := r.name()
	for _, attached := range task.Attached {
		if attached.Name == name {
			found = true
			err = r.DB.First(&file, attached.ID).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	if !found {
		return
	}
	f, err = os.OpenFile(file.Path, os.O_RDONLY|os.O_APPEND, 0666)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	st, err := f.Stat()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.nSkip = st.Size()
	return
}

// create creates and attaches the log file.
func (r *LogCollector) create(task *Task) (f *os.File, err error) {
	file := &model.File{Name: r.name()}
	err = r.DB.Create(file).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	f, err = os.Create(file.Path)
	if err != nil {
		_ = r.DB.Delete(file)
		err = liberr.Wrap(err)
		return
	}
	task.attach(file)
	return
}

// copy data.
// The read bytes are discarded when smaller than nSkip.
// The offset is adjusted when to account for the buffer
// containing bytes to be skipped and written.
func (r *LogCollector) copy(reader io.Reader, writer io.Writer) (err error) {
	if r.nBuf < 1 {
		r.nBuf = 0x8000
	}
	buf := make([]byte, r.nBuf)
	for {
		n, rErr := reader.Read(buf)
		if rErr != nil {
			if rErr != io.EOF {
				err = rErr
			}
			break
		}
		nRead := int64(n)
		if nRead == 0 {
			continue
		}
		offset := int64(0)
		if r.nSkip > 0 {
			if nRead > r.nSkip {
				offset = r.nSkip
				r.nSkip = 0
			} else {
				r.nSkip -= nRead
				continue
			}
		}
		b := buf[offset:nRead]
		_, err = writer.Write(b)
		if err != nil {
			return
		}
		if f, cast := writer.(*os.File); cast {
			err = f.Sync()
			if err != nil {
				return
			}
		}
	}
	return
}
