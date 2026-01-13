package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Task API.
type Task struct {
	client *Client
}

// Create a Task.
func (h *Task) Create(r *api2.Task) (err error) {
	err = h.client.Post(api2.TasksRoute, &r)
	return
}

// Get a Task by ID.
func (h *Task) Get(id uint) (r *api2.Task, err error) {
	r = &api2.Task{}
	path := Path(api2.TaskRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tasks.
func (h *Task) List() (list []api2.Task, err error) {
	list = []api2.Task{}
	err = h.client.Get(api2.TasksRoute, &list)
	return
}

// BulkCancel - Cancel tasks matched by filter.
func (h *Task) BulkCancel(filter Filter) (err error) {
	err = h.client.Put(api2.TasksCancelRoute, 0, filter.Param())
	return
}

// Update a Task.
func (h *Task) Update(r *api2.Task) (err error) {
	path := Path(api2.TaskRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Patch a Task.
func (h *Task) Patch(id uint, r any) (err error) {
	path := Path(api2.TaskRoute).Inject(Params{api2.ID: id})
	err = h.client.Patch(path, r)
	return
}

// Delete a Task.
func (h *Task) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.TaskRoute).Inject(Params{api2.ID: id}))
	return
}

// Bucket returns the bucket API.
func (h *Task) Bucket(id uint) (b *BucketContent) {
	params := Params{
		api2.ID:       id,
		api2.Wildcard: "",
	}
	path := Path(api2.TaskBucketContentRoute).Inject(params)
	b = &BucketContent{
		root:   path,
		client: h.client,
	}
	return
}
