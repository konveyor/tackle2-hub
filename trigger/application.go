package trigger

import (
	"fmt"
	"sort"

	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
)

// Application trigger.
type Application struct {
	Trigger
}

// Created trigger.
func (r *Application) Created(m *model.Application) (err error) {
	err = r.Updated(m)
	return
}

// Updated trigger.
func (r *Application) Updated(m *model.Application) (err error) {
	if !Settings.Discovery.Enabled {
		return
	}
	if m.Repository == (model.Repository{}) {
		return
	}
	label := Settings.Discovery.Label
	kinds, err := r.FindTasks(label)
	if err != nil {
		return
	}
	sort.Slice(
		kinds,
		func(i, j int) bool {
			ik := kinds[i]
			jk := kinds[j]
			iL := ik.Labels[label]
			jL := jk.Labels[label]
			return iL < jL
		})
	taskGroup := &tasking.TaskGroup{
		TaskGroup: &model.TaskGroup{
			Mode: tasking.Pipeline,
		},
	}
	taskGroup.Mode = tasking.Pipeline
	for _, kind := range kinds {
		task := model.Task{}
		task.Kind = kind.Name
		task.Name = fmt.Sprintf("%s-%s", m.Name, kind.Name)
		task.ApplicationID = &m.ID
		taskGroup.List =
			append(
				taskGroup.List,
				task)
	}
	err = taskGroup.Submit(r.DB, r.TaskManager)
	return
}
