package trigger

import (
	"fmt"

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
	kinds, err := r.FindTasks(Settings.Discovery.Label)
	if err != nil {
		return
	}
	for _, kind := range kinds {
		t := &tasking.Task{Task: &model.Task{}}
		t.Kind = kind.Name
		t.Name = fmt.Sprintf("%s-%s", m.Name, t.Name)
		t.ApplicationID = &m.ID
		t.State = tasking.Ready
		err = r.TaskManager.Create(r.DB, t)
		if err != nil {
			return
		}
	}
	return
}
