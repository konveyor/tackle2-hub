package resource

import (
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	tasking "github.com/konveyor/tackle2-hub/task"
)

// TaskGroup REST resource.
type TaskGroup api.TaskGroup

// With updates the resource with the model.
func (r *TaskGroup) With(m *model.TaskGroup) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.Extensions = m.Extensions
	r.State = m.State
	r.Priority = m.Priority
	r.Policy = api.TaskPolicy(m.Policy)
	r.Data = m.Data.Any
	r.Bucket = refPtr(m.BucketID, m.Bucket)
	r.Tasks = []api.Task{}
	switch m.State {
	case "", tasking.Created:
		for _, task := range m.List {
			member := Task{}
			member.With(&task)
			r.Tasks = append(
				r.Tasks,
				api.Task(member))
		}
	default:
		for _, task := range m.Tasks {
			member := Task{}
			member.With(&task)
			r.Tasks = append(
				r.Tasks,
				api.Task(member))
		}
	}
}

// Model builds a model.
func (r *TaskGroup) Model() (m *model.TaskGroup) {
	m = &model.TaskGroup{}
	m.ID = r.ID
	m.Name = r.Name
	m.Kind = r.Kind
	m.Addon = r.Addon
	m.Extensions = r.Extensions
	m.State = r.State
	m.Priority = r.Priority
	m.Policy = model.TaskPolicy(r.Policy)
	m.Data.Any = r.Data
	m.List = make([]model.Task, 0, len(r.Tasks))
	for _, task := range r.Tasks {
		t := Task(task)
		tm := t.Model()
		m.List = append(m.List, *tm)
	}
	if r.Bucket != nil {
		m.BucketID = &r.Bucket.ID
	}
	return
}
