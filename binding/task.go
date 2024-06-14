package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Task API.
type Task struct {
	client *Client
}

// Create a Task.
func (h *Task) Create(r *api.Task) (err error) {
	err = h.client.Post(api.TasksRoot, &r)
	return
}

// Get a Task by ID.
func (h *Task) Get(id uint) (r *api.Task, err error) {
	r = &api.Task{}
	path := Path(api.TaskRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tasks.
func (h *Task) List() (list []api.Task, err error) {
	list = []api.Task{}
	err = h.client.Get(api.TasksRoot, &list)
	return
}

// Update a Task.
func (h *Task) Update(r *api.Task) (err error) {
	path := Path(api.TaskRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Patch a Task.
func (h *Task) Patch(r *api.Task) (err error) {
	path := Path(api.TaskRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Patch(path, r)
	return
}

// Delete a Task.
func (h *Task) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TaskRoot).Inject(Params{api.ID: id}))
	return
}

// Bucket returns the bucket API.
func (h *Task) Bucket(id uint) (b *BucketContent) {
	params := Params{
		api.ID:       id,
		api.Wildcard: "",
	}
	path := Path(api.TaskBucketContentRoot).Inject(params)
	b = &BucketContent{
		root:   path,
		client: h.client,
	}
	return
}
