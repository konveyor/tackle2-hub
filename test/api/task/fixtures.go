package task

import (
	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = c.Client
)

//
// Set of valid resources for tests and reuse.
func Samples() (samples []api.Task) {
	samples = []api.Task{
		{
			Name:  "Test windup task",
			Addon: "windup",
			Data:  "{}",
		},
	}

	return
}

//
// Create a Task.
func Create(r *api.Task) (err error) {
	err = Client.Post(api.TasksRoot, &r)
	return
}

//
// Retrieve the Task.
func Get(r *api.Task) (err error) {
	err = Client.Get(c.Path(api.TaskRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Update the Task.
func Update(r *api.Task) (err error) {
	err = Client.Put(c.Path(api.TaskRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Delete the Task.
func Delete(r *api.Task) (err error) {
	err = Client.Delete(c.Path(api.TaskRoot, c.Params{api.ID: r.ID}))
	return
}

//
// List Tasks.
func List(r []*api.Task) (err error) {
	err = Client.Get(api.TasksRoot, &r)
	return
}
