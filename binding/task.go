package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Task API.
type Task struct {
	// hub API client.
	Client *Client
}

//
// Create a Task.
func (h *Task) Create(r *api.Task) (err error) {
	err = h.Client.Post(api.TasksRoot, &r)
	return
}

//
// Get a Task by ID.
func (h *Task) Get(id uint) (r *api.Task, err error) {
	r = &api.Task{}
	path := Path(api.TaskRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List Tasks.
func (h *Task) List() (list []api.Task, err error) {
	list = []api.Task{}
	err = h.Client.Get(api.TasksRoot, &list)
	return
}

//
// Update a Task.
func (h *Task) Update(r *api.Task) (err error) {
	path := Path(api.TaskRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete a Task.
func (h *Task) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.TaskRoot).Inject(Params{api.ID: id}))
	return
}
