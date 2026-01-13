package resource

import (
	"io/ioutil"
	"sort"
	"strings"

	"github.com/konveyor/tackle2-hub/internal/model"
	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm"
	"k8s.io/utils/strings/slices"
)

// TTL REST resource.
type TTL = api.TTL

// TaskPolicy REST resource.
type TaskPolicy = api.TaskPolicy

// TaskError REST resource.
type TaskError = api.TaskError

// TaskEvent REST resource.
type TaskEvent = api.TaskEvent

// Attachment REST resource.
type Attachment = api.Attachment

// Task REST resource.
type Task api.Task

// With updates the resource with the model.
func (r *Task) With(m *model.Task) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.Extensions = m.Extensions
	r.State = m.State
	r.Locator = m.Locator
	r.Priority = r.userPriority(m.Priority)
	r.Policy = TaskPolicy(m.Policy)
	r.TTL = TTL(m.TTL)
	r.Data = m.Data.Any
	r.Application = refPtr(m.ApplicationID, m.Application)
	r.Platform = refPtr(m.PlatformID, m.Platform)
	r.Bucket = refPtr(m.BucketID, m.Bucket)
	r.Pod = m.Pod
	r.Retries = m.Retries
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Events = make([]TaskEvent, 0)
	r.Errors = make([]TaskError, 0)
	r.Attached = make([]Attachment, 0)
	for _, event := range m.Events {
		r.Events = append(r.Events, TaskEvent(event))
	}
	for _, err := range m.Errors {
		r.Errors = append(r.Errors, TaskError(err))
	}
	if m.Report != nil {
		report := &TaskReport{}
		report.With(m.Report)
		r.Activity = report.Activity
		r.Errors = append(r.Errors, report.Errors...)
		r.Attached = append(r.Attached, report.Attached...)
		switch r.State {
		case tasking.Succeeded:
			switch report.Status {
			case tasking.Failed:
				r.State = report.Status
			}
		}
	}
	for _, a := range m.Attached {
		r.Attached = append(r.Attached, Attachment(a))
	}
}

// Patch the specified model.
func (r *Task) Patch(m *model.Task) {
	m.ID = r.ID
	m.Name = r.Name
	m.Kind = r.Kind
	m.Addon = r.Addon
	m.Extensions = r.Extensions
	m.State = r.State
	m.Locator = r.Locator
	m.Priority = r.Priority
	m.Policy = model.TaskPolicy(r.Policy)
	m.TTL = model.TTL(r.TTL)
	m.Data.Any = r.Data
	m.ApplicationID = idPtr(r.Application)
	m.PlatformID = idPtr(r.Platform)
}

// InjectFiles inject attached files into the activity.
func (r *Task) InjectFiles(db *gorm.DB) (err error) {
	sort.Slice(
		r.Attached,
		func(i, j int) bool {
			return r.Attached[i].Activity > r.Attached[j].Activity
		})
	for _, ref := range r.Attached {
		if ref.Activity == 0 {
			continue
		}
		if ref.Activity > len(r.Activity) {
			continue
		}
		m := &model.File{}
		err = db.First(m, ref.ID).Error
		if err != nil {
			return
		}
		b, nErr := ioutil.ReadFile(m.Path)
		if nErr != nil {
			err = nErr
			return
		}
		var content []string
		for _, s := range strings.Split(string(b), "\n") {
			content = append(
				content,
				"> "+s)
		}
		snipA := slices.Clone(r.Activity[:ref.Activity])
		snipB := slices.Clone(r.Activity[ref.Activity:])
		r.Activity = append(
			append(snipA, content...),
			snipB...)
	}
	return
}

// userPriority adjust (ensures) priority is greater than 10.
// Priority: 0-9 reserved for system tasks.
func (r *Task) userPriority(in int) (out int) {
	out = in
	if out < 10 {
		out += 10
	}
	return
}

// TaskReport REST resource.
type TaskReport api.TaskReport

// With updates the resource with the model.
func (r *TaskReport) With(m *model.TaskReport) {
	baseWith(&r.Resource, &m.Model)
	r.Status = m.Status
	r.Total = m.Total
	r.Completed = m.Completed
	r.TaskID = m.TaskID
	r.Activity = m.Activity
	r.Result = m.Result.Any
	r.Errors = make([]TaskError, 0, len(m.Errors))
	r.Attached = make([]Attachment, 0, len(m.Attached))
	for _, err := range m.Errors {
		r.Errors = append(r.Errors, TaskError(err))
	}
	for _, a := range m.Attached {
		r.Attached = append(r.Attached, Attachment(a))
	}
}

// Patch the specified model.
func (r *TaskReport) Patch(m *model.TaskReport) {
	m.ID = r.ID
	m.Status = r.Status
	m.Total = r.Total
	m.Completed = r.Completed
	m.Activity = r.Activity
	m.TaskID = r.TaskID
	m.Result.Any = r.Result
	m.Errors = make([]model.TaskError, 0, len(r.Errors))
	m.Attached = make([]model.Attachment, 0, len(r.Attached))
	for _, err := range r.Errors {
		m.Errors = append(m.Errors, model.TaskError(err))
	}
	for _, at := range r.Attached {
		m.Attached = append(m.Attached, model.Attachment(at))
	}
}

// TaskQueue REST resource.
type TaskQueue = api.TaskQueue

// TaskDashboard REST resource.
type TaskDashboard api.TaskDashboard

func (r *TaskDashboard) With(m *model.Task) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.State = m.State
	r.Locator = m.Locator
	r.Application = refPtr(m.ApplicationID, m.Application)
	r.Platform = refPtr(m.PlatformID, m.Platform)
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Errors = len(m.Errors)
	if m.Report != nil {
		r.Errors += len(m.Report.Errors)
	}
}
