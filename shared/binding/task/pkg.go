package task

import (
	"path/filepath"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) (h Task) {
	h = Task{client: client}
	return
}

// Task API.
type Task struct {
	client client.RestClient
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

// Submit a Task.
func (h Task) Submit(id uint) (err error) {
	path := client.Path(api.TaskSubmitRoute).Inject(client.Params{api.ID: id})
	err = h.client.Put(path, nil)
	return
}

// Cancel a Task.
func (h Task) Cancel(id uint) (err error) {
	path := client.Path(api.TaskCancelRoute).Inject(client.Params{api.ID: id})
	err = h.client.Put(path, nil)
	return
}

// GetAttached downloads the attached resources for a Task as a tarball.
func (h Task) GetAttached(id uint, destination string) (err error) {
	path := client.Path(api.TaskAttachedRoute).Inject(client.Params{api.ID: id})
	isDir, err := h.client.IsDir(destination, false)
	if err != nil {
		return
	}
	if isDir {
		r := &api.File{}
		err = h.client.Get(path, r)
		if err != nil {
			return
		}
		destination = filepath.Join(
			destination,
			r.Name)
	}
	err = h.client.FileGet(path, destination)
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
