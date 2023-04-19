package task

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Set of valid resources for tests and reuse.
func Samples() (samples map[string]api.Task) {
	samples = map[string]api.Task{
		"Windup": {
			Name:  "Test windup task",
			Addon: "windup",
			Data:  "{}",
		},
	}

	return
}

// Create a Task.
func Create(r *api.Task) (err error) {
	err = Client.Post(api.TasksRoot, &r)
	return
}

// Retrieve the Task.
func Get(r *api.Task) (err error) {
	err = Client.Get(client.Path(api.TaskRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the Task.
func Update(r *api.Task) (err error) {
	err = Client.Put(client.Path(api.TaskRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the Task.
func Delete(r *api.Task) (err error) {
	err = Client.Delete(client.Path(api.TaskRoot, client.Params{api.ID: r.ID}))
	return
}

// List Tasks.
func List(r []*api.Task) (err error) {
	err = Client.Get(api.TasksRoot, &r)
	return
}
