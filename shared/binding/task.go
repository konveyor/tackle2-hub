package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Task API.
type Task struct {
	client *Client
}

// Create a Task.
func (h *Task) Create(r *api.Task) (err error) {
	err = h.client.Post(api.TasksRoute, &r)
	return
}

// Get a Task by ID.
func (h *Task) Get(id uint) (r *api.Task, err error) {
	r = &api.Task{}
	path := Path(api.TaskRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tasks.
func (h *Task) List() (list []api.Task, err error) {
	list = []api.Task{}
	err = h.client.Get(api.TasksRoute, &list)
	return
}

// BulkCancel - Cancel tasks matched by filter.
func (h *Task) BulkCancel(filter Filter) (err error) {
	err = h.client.Put(api.TasksCancelRoute, 0, filter.Param())
	return
}

// Update a Task.
func (h *Task) Update(r *api.Task) (err error) {
	path := Path(api.TaskRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Patch a Task.
func (h *Task) Patch(id uint, r any) (err error) {
	path := Path(api.TaskRoute).Inject(Params{api.ID: id})
	err = h.client.Patch(path, r)
	return
}

// Delete a Task.
func (h *Task) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TaskRoute).Inject(Params{api.ID: id}))
	return
}

// Bucket returns the bucket API.
func (h *Task) Bucket(id uint) (b *BucketContent) {
	params := Params{
		api.ID:       id,
		api.Wildcard: "",
	}
	path := Path(api.TaskBucketContentRoute).Inject(params)
	b = &BucketContent{
		root:   path,
		client: h.client,
	}
	return
}
