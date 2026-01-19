package task

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client *client.Client) (h Task) {
	h = Task{client: client}
	return
}

// Task API.
type Task struct {
	client *client.Client
}

// Create a Task.
func (h Task) Create(r *api.Task) (err error) {
	err = h.client.Post(api.TasksRoute, r)
	return
}

// Get a Task by ID.
func (h Task) Get(id uint) (r *api.Task, err error) {
	r = &api.Task{}
	path := client.Path(api.TaskRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tasks.
func (h Task) List() (list []api.Task, err error) {
	list = []api.Task{}
	err = h.client.Get(api.TasksRoute, &list)
	return
}

// BulkCancel - Cancel tasks matched by filter.
func (h Task) BulkCancel(filter client.Filter) (err error) {
	err = h.client.Put(api.TasksCancelRoute, 0, filter.Param())
	return
}

// Update a Task.
func (h Task) Update(r *api.Task) (err error) {
	path := client.Path(api.TaskRoute).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Patch a Task.
func (h Task) Patch(id uint, r any) (err error) {
	path := client.Path(api.TaskRoute).Inject(client.Params{api.ID: id})
	err = h.client.Patch(path, r)
	return
}

// Delete a Task.
func (h Task) Delete(id uint) (err error) {
	path := client.Path(api.TaskRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// Bucket returns the bucket API.
// Deprecated. Use Select().
func (h Task) Bucket(id uint) (h2 bucket.Content) {
	selected := h.Select(id)
	h2 = selected.Bucket
	return
}

// Select returns the API for the selected task.
func (h Task) Select(id uint) (h2 Selected) {
	h2 = Selected{}
	path := client.Path(api.TaskBucketContentRoute).
		Inject(client.Params{
			api.ID:       id,
			api.Wildcard: "",
		})
	h2.Bucket = bucket.NewContent(h.client, path)
	return
}

// Selected task API.
type Selected struct {
	Bucket bucket.Content
}
