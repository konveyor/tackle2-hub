package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/auth"
	k8s2 "github.com/konveyor/tackle2-hub/k8s"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha2"
	"github.com/konveyor/tackle2-hub/metrics"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
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
	AddonSelected = "AddonSelected"
	ExtSelected   = "ExtensionSelected"
	ImageError    = "ImageError"
	PodNotFound   = "PodNotFound"
	PodCreated    = "PodCreated"
	PodPending    = "PodPending"
	PodRunning    = "PodRunning"
	Preempted     = "Preempted"
	PodSucceeded  = "PodSucceeded"
	PodFailed     = "PodFailed"
	PodDeleted    = "PodDeleted"
	Escalated     = "Escalated"
	Released      = "Released"
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
	Shared = "shared"
	Cache  = "cache"
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
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	m.queue = make(chan func(), 100)
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
					m.deleteOrphanPods()
					m.runActions()
					m.updateRunning()
					m.startReady()
					m.pause()
				} else {
					Log.Error(err, "")
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
func (m *Manager) Update(db *gorm.DB, task *Task) (err error) {
	found := &Task{}
	err = db.First(found, task.ID).Error
	if err != nil {
		return
	}
	switch task.State {
	case Created:
		db = db.Select(
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
		err = m.findRefs(task)
		if err != nil {
			return
		}
		db = db.Where("State", Created)
		err = db.Save(task).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case Ready,
		Pending,
		QuotaBlocked,
		Postponed:
		db = db.Select(
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
		err = db.Save(task).Error
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
		if extension.Spec.Addon != addon.Name {
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
		if !task.StateIn(Running, Pending) {
			continue
		}
		running := task
		pod, found := running.Reflect(&m.cluster)
		if found {
			if task.StateIn(Succeeded, Failed) {
				err = m.podSnapshot(running, pod)
				if err != nil {
					Log.Error(err, "")
					continue
				}
				err = running.Delete(m.Client)
				if err != nil {
					Log.Error(err, "")
					continue
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
//   - container Logs
func (m *Manager) podSnapshot(task *Task, pod *core.Pod) (err error) {
	var files []*model.File
	d, err := m.podYAML(pod)
	if err != nil {
		return
	}
	files = append(files, d)
	logs, err := m.podLogs(pod)
	if err != nil {
		return
	}
	files = append(files, logs...)
	for _, f := range files {
		task.attach(f)
	}
	Log.V(1).Info("Task pod snapshot attached.", "id", task.ID)
	return
}

// podYAML builds pod resource description.
func (m *Manager) podYAML(pod *core.Pod) (file *model.File, err error) {
	events, err := m.podEvent(pod)
	if err != nil {
		return
	}
	file = &model.File{Name: "pod.yaml"}
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

// podLogs - get and store pod logs as a Files.
func (m *Manager) podLogs(pod *core.Pod) (files []*model.File, err error) {
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil {
			continue
		}
		f, nErr := m.containerLog(pod, container.Name)
		if nErr == nil {
			files = append(files, f)
		} else {
			err = nErr
			return
		}
	}
	return
}

// containerLog - get container log and store in file.
func (m *Manager) containerLog(pod *core.Pod, container string) (file *model.File, err error) {
	options := &core.PodLogOptions{
		Container: container,
	}
	clientSet, err := k8s2.NewClientSet()
	if err != nil {
		return
	}
	podClient := clientSet.CoreV1().Pods(Settings.Hub.Namespace)
	req := podClient.GetLogs(pod.Name, options)
	reader, err := req.Stream(context.TODO())
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	file = &model.File{Name: container + ".log"}
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
	_, err = io.Copy(f, reader)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Task is an runtime task.
type Task struct {
	// model.
	*model.Task
}

func (r *Task) With(m *model.Task) {
	r.Task = m
}

// StateIn returns true matches on of the specified states.
func (r *Task) StateIn(states ...string) (matched bool) {
	for _, state := range states {
		if r.State == state {
			matched = true
			break
		}
	}
	return
}

// Error appends an error.
func (r *Task) Error(severity, description string, x ...any) {
	description = fmt.Sprintf(description, x...)
	r.Errors = append(
		r.Errors,
		model.TaskError{
			Severity:    severity,
			Description: description,
		})
}

// Event appends an event.
// Duplicates result in count incremented and Last updated.
func (r *Task) Event(kind string, p ...any) {
	mark := time.Now()
	reason := ""
	if len(p) > 0 {
		switch x := p[0].(type) {
		case string:
			reason = fmt.Sprintf(x, p[1:]...)
		case int:
			reason = strconv.Itoa(x)
		}
	}
	event, found := r.LastEvent(kind)
	if found && event.Reason == reason {
		event.Last = mark
		event.Count++
		return
	}
	event = &model.TaskEvent{
		Kind:   kind,
		Count:  1,
		Reason: reason,
		Last:   mark,
	}
	r.Events = append(r.Events, *event)
}

// LastEvent returns the last event of the specified kind.
func (r *Task) LastEvent(kind string) (event *model.TaskEvent, found bool) {
	for i := len(r.Events) - 1; i >= 0; i-- {
		event = &r.Events[i]
		if kind == event.Kind {
			found = true
			break
		}
	}
	return
}

// Run the specified task.
func (r *Task) Run(cluster *Cluster) (started bool, err error) {
	mark := time.Now()
	client := cluster.Client
	defer func() {
		if err == nil {
			return
		}
		matched, retry := SoftErr(err)
		if matched {
			if !retry {
				r.Error("Error", err.Error())
				r.Terminated = &mark
				r.State = Failed
			}
			err = nil
		}
	}()
	addon, found := cluster.Addon(r.Addon)
	if !found {
		err = &AddonNotFound{Name: r.Addon}
		return
	}
	extensions, err := r.getExtensions(client)
	if err != nil {
		return
	}
	for _, extension := range extensions {
		if r.Addon != extension.Spec.Addon {
			err = &ExtensionNotValid{
				Name:  extension.Name,
				Addon: addon.Name,
			}
			return
		}
	}
	secret := r.secret()
	err = client.Create(context.TODO(), &secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		if err != nil {
			_ = client.Delete(context.TODO(), &secret)
		}
	}()
	pod := r.pod(
		addon,
		extensions,
		cluster.Tackle(),
		&secret)
	err = client.Create(context.TODO(), &pod)
	if err != nil {
		qe := &QuotaExceeded{}
		if qe.Match(err) {
			r.State = QuotaBlocked
			r.Event(QuotaBlocked, qe.Reason)
			err = qe
			return
		}
		pe := &PodRejected{}
		if pe.Match(err) {
			err = liberr.Wrap(pe)
			return
		}
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		if err != nil {
			_ = client.Delete(context.TODO(), &pod)
		}
	}()
	secret.OwnerReferences = append(
		secret.OwnerReferences,
		meta.OwnerReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       pod.Name,
			UID:        pod.UID,
		})
	err = client.Update(context.TODO(), &secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	started = true
	r.Started = &mark
	r.State = Pending
	r.Pod = path.Join(
		pod.Namespace,
		pod.Name)
	r.Event(PodCreated, r.Pod)
	return
}

// Reflect finds the associated pod and updates the task state.
func (r *Task) Reflect(cluster *Cluster) (pod *core.Pod, found bool) {
	pod, found = cluster.Pod(path.Base(r.Pod))
	if !found {
		r.State = Ready
		r.Event(PodNotFound, r.Pod)
		r.Terminated = nil
		r.Started = nil
		r.Pod = ""
		return
	}
	client := cluster.Client
	switch pod.Status.Phase {
	case core.PodPending:
		r.podPending(pod)
	case core.PodRunning:
		r.podRunning(pod, client)
	case core.PodSucceeded:
		r.podSucceeded(pod)
	case core.PodFailed:
		r.podFailed(pod, client)
	}

	return
}

// Delete the associated pod as needed.
func (r *Task) Delete(client k8s.Client) (err error) {
	if r.Pod == "" {
		return
	}
	pod := &core.Pod{}
	pod.Namespace = path.Dir(r.Pod)
	pod.Name = path.Base(r.Pod)
	err = client.Delete(context.TODO(), pod, k8s.GracePeriodSeconds(0))
	if err != nil {
		if !k8serr.IsNotFound(err) {
			err = liberr.Wrap(err)
			return
		} else {
			err = nil
		}
	}
	r.Pod = ""
	r.Event(PodDeleted, r.Pod)
	Log.Info(
		"Task pod deleted.",
		"id",
		r.ID,
		"pod",
		pod.Name)
	mark := time.Now()
	r.Terminated = &mark
	return
}

// podPending handles pod pending.
func (r *Task) podPending(pod *core.Pod) {
	var status []core.ContainerStatus
	status = append(
		status,
		pod.Status.InitContainerStatuses...)
	status = append(
		status,
		pod.Status.ContainerStatuses...)
	started := 0
	for _, cnd := range pod.Status.Conditions {
		if cnd.Type == core.PodScheduled &&
			cnd.Reason == core.PodReasonUnschedulable {
			r.Event(PodPending, cnd.Message)
			return
		}
	}
	for _, status := range status {
		state := status.State
		if state.Waiting != nil {
			waiting := state.Waiting
			reason := strings.ToLower(waiting.Reason)
			if r.containsAny(reason, "invalid", "error", "backoff") {
				r.Error(
					"Error",
					"Container (%s) failed: %s",
					status.Name,
					waiting.Reason)
				mark := time.Now()
				r.Terminated = &mark
				r.Event(ImageError, waiting.Reason)
				r.State = Failed
				return
			} else {
				r.Event(PodPending, waiting.Reason)
			}
		}
		if status.Started == nil {
			continue
		}
		if *status.Started {
			started++
		}
	}
	if started > 0 {
		r.Event(PodRunning)
		r.State = Running
	}
}

// Cancel the task.
func (r *Task) Cancel(client k8s.Client) (err error) {
	err = r.Delete(client)
	if err != nil {
		return
	}
	r.State = Canceled
	r.Event(Canceled)
	r.SetBucket(nil)
	Log.Info(
		"Task canceled.",
		"id",
		r.ID)
	return
}

// podRunning handles pod running.
func (r *Task) podRunning(pod *core.Pod, client k8s.Client) {
	r.State = Running
	r.Event(PodRunning)
	addonStatus := pod.Status.ContainerStatuses[0]
	if addonStatus.State.Terminated != nil {
		switch addonStatus.State.Terminated.ExitCode {
		case 0:
			r.podSucceeded(pod)
		default: // failed.
			r.podFailed(pod, client)
			return
		}
	}
}

// podFailed handles pod succeeded.
func (r *Task) podSucceeded(pod *core.Pod) {
	mark := time.Now()
	r.State = Succeeded
	r.Event(PodSucceeded)
	r.Terminated = &mark
}

// podFailed handles pod failed.
func (r *Task) podFailed(pod *core.Pod, client k8s.Client) {
	mark := time.Now()
	var statuses []core.ContainerStatus
	statuses = append(
		statuses,
		pod.Status.InitContainerStatuses...)
	statuses = append(
		statuses,
		pod.Status.ContainerStatuses...)
	for _, status := range statuses {
		if status.State.Terminated == nil {
			continue
		}
		r.Event(
			PodFailed,
			status.State.Terminated.Reason)
		switch status.State.Terminated.ExitCode {
		case 0: // Succeeded.
		case 137: // Killed.
			if r.Retries < Settings.Hub.Task.Retries {
				_ = client.Delete(context.TODO(), pod)
				r.Pod = ""
				r.State = Ready
				r.Terminated = nil
				r.Started = nil
				r.Errors = nil
				r.Retries++
				return
			}
			fallthrough
		default: // Error.
			r.State = Failed
			r.Terminated = &mark
			r.Error(
				"Error",
				"Container (%s) failed: %s",
				status.Name,
				status.State.Terminated.Reason)
			return
		}
	}
}

// getExtensions by name.
func (r *Task) getExtensions(client k8s.Client) (extensions []crd.Extension, err error) {
	for _, name := range r.Extensions {
		extension := crd.Extension{}
		err = client.Get(
			context.TODO(),
			k8s.ObjectKey{
				Namespace: Settings.Hub.Namespace,
				Name:      name,
			},
			&extension)
		if err != nil {
			if k8serr.IsNotFound(err) {
				err = &ExtensionNotFound{name}
			} else {
				err = liberr.Wrap(err)
			}
			return
		}
		extensions = append(
			extensions,
			extension)
	}
	return
}

// pod build the pod.
func (r *Task) pod(
	addon *crd.Addon,
	extensions []crd.Extension,
	owner *crd.Tackle,
	secret *core.Secret) (pod core.Pod) {
	//
	pod = core.Pod{
		Spec: r.specification(addon, extensions, secret),
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: r.k8sName(),
			Labels:       r.labels(),
		},
	}
	pod.OwnerReferences = append(
		pod.OwnerReferences,
		meta.OwnerReference{
			APIVersion: owner.APIVersion,
			Kind:       owner.Kind,
			Name:       owner.Name,
			UID:        owner.UID,
		})
	return
}

// specification builds a Pod specification.
func (r *Task) specification(
	addon *crd.Addon,
	extensions []crd.Extension,
	secret *core.Secret) (specification core.PodSpec) {
	shared := core.Volume{
		Name: Shared,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	cache := core.Volume{
		Name: Cache,
	}
	if Settings.Cache.RWX {
		cache.VolumeSource = core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: Settings.Cache.PVC,
			},
		}
	} else {
		cache.VolumeSource = core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		}
	}
	init, plain := r.containers(addon, extensions, secret)
	specification = core.PodSpec{
		ServiceAccountName: Settings.Hub.Task.SA,
		RestartPolicy:      core.RestartPolicyNever,
		InitContainers:     init,
		Containers:         plain,
		Volumes: []core.Volume{
			shared,
			cache,
		},
	}

	return
}

// container builds the pod containers.
func (r *Task) containers(
	addon *crd.Addon,
	extensions []crd.Extension,
	secret *core.Secret) (init []core.Container, plain []core.Container) {
	userid := int64(0)
	token := &core.EnvVarSource{
		SecretKeyRef: &core.SecretKeySelector{
			Key: settings.EnvHubToken,
			LocalObjectReference: core.LocalObjectReference{
				Name: secret.Name,
			},
		},
	}
	plain = append(plain, addon.Spec.Container)
	plain[0].Name = "addon"
	for i := range extensions {
		extension := &extensions[i]
		container := extension.Spec.Container
		container.Name = extension.Name
		plain = append(
			plain,
			container)
	}
	injector := Injector{}
	for i := range plain {
		container := &plain[i]
		injector.Inject(container)
		r.propagateEnv(&plain[0], container)
		container.SecurityContext = &core.SecurityContext{
			RunAsUser: &userid,
		}
		container.VolumeMounts = append(
			container.VolumeMounts,
			core.VolumeMount{
				Name:      Shared,
				MountPath: Settings.Shared.Path,
			},
			core.VolumeMount{
				Name:      Cache,
				MountPath: Settings.Cache.Path,
			})
		container.Env = append(
			container.Env,
			core.EnvVar{
				Name:  settings.EnvSharedPath,
				Value: Settings.Shared.Path,
			},
			core.EnvVar{
				Name:  settings.EnvCachePath,
				Value: Settings.Cache.Path,
			},
			core.EnvVar{
				Name:  settings.EnvHubBaseURL,
				Value: Settings.Addon.Hub.URL,
			},
			core.EnvVar{
				Name:  settings.EnvTask,
				Value: strconv.Itoa(int(r.Task.ID)),
			},
			core.EnvVar{
				Name:      settings.EnvHubToken,
				ValueFrom: token,
			})
	}
	return
}

// propagateEnv copies extension container Env.* to the addon container.
// Prefixed with EXTENSION_<name>.
func (r *Task) propagateEnv(addon, extension *core.Container) {
	for _, env := range extension.Env {
		addon.Env = append(
			addon.Env,
			core.EnvVar{
				Name:  ExtEnv(extension.Name, env.Name),
				Value: env.Value,
			})
	}
}

// secret builds the pod secret.
func (r *Task) secret() (secret core.Secret) {
	user := "addon:" + r.Addon
	token, _ := auth.Hub.NewToken(
		user,
		auth.AddonRole,
		jwt.MapClaims{
			"task": r.ID,
		})
	secret = core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: r.k8sName(),
			Labels:       r.labels(),
		},
		Data: map[string][]byte{
			settings.EnvHubToken: []byte(token),
		},
	}

	return
}

// k8sName returns a name suitable to be used for k8s resources.
func (r *Task) k8sName() string {
	return fmt.Sprintf("task-%d-", r.ID)
}

// labels builds k8s labels.
func (r *Task) labels() map[string]string {
	return map[string]string{
		TaskLabel: strconv.Itoa(int(r.ID)),
		AppLabel:  "tackle",
		RoleLabel: "task",
	}
}

// attach file.
func (r *Task) attach(file *model.File) {
	r.Attached = append(
		r.Attached,
		model.Attachment{
			ID:   file.ID,
			Name: file.Name,
		})
}

// containsAny returns true when the str contains any of substr.
func (r *Task) containsAny(str string, substr ...string) (matched bool) {
	for i := range substr {
		if strings.Contains(str, substr[i]) {
			matched = true
			break
		}
	}
	return
}

// update manager controlled fields.
func (r *Task) update(db *gorm.DB) (err error) {
	db = db.Select(
		"Addon",
		"Extensions",
		"State",
		"Priority",
		"Started",
		"Terminated",
		"Events",
		"Error",
		"Retries",
		"Attached",
		"Pod")
	err = db.Save(r).Error
	return
}

// Event represents a pod event.
type Event struct {
	Type     string
	Reason   string
	Age      string
	Reporter string
	Message  string
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
