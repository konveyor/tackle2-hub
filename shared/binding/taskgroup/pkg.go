package taskgroup

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) (h TaskGroup) {
	h = TaskGroup{client: client}
	return
}

// TaskGroup API.
type TaskGroup struct {
	client client.RestClient
}

// Create a TaskGroup.
func (h TaskGroup) Create(r *api.TaskGroup) (err error) {
	err = h.client.Post(api.TaskGroupsRoute, r)
	return
}

// Get a TaskGroup by ID.
func (h TaskGroup) Get(id uint) (r *api.TaskGroup, err error) {
	r = &api.TaskGroup{}
	path := client.Path(api.TaskGroupRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List TaskGroups.
func (h TaskGroup) List() (list []api.TaskGroup, err error) {
	list = []api.TaskGroup{}
	err = h.client.Get(api.TaskGroupsRoute, &list)
	return
}

// Update a TaskGroup.
func (h TaskGroup) Update(r *api.TaskGroup) (err error) {
	path := client.Path(api.TaskGroupRoute).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Patch a TaskGroup.
func (h TaskGroup) Patch(id uint, r any) (err error) {
	path := client.Path(api.TaskGroupRoute).Inject(client.Params{api.ID: id})
	err = h.client.Patch(path, r)
	return
}

// Delete a TaskGroup.
func (h TaskGroup) Delete(id uint) (err error) {
	path := client.Path(api.TaskGroupRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// Submit a TaskGroup.
func (h TaskGroup) Submit(id uint) (err error) {
	path := client.Path(api.TaskGroupSubmitRoute).Inject(client.Params{api.ID: id})
	err = h.client.Put(path, nil)
	return
}

// Select returns the API for the selected task group.
func (h TaskGroup) Select(id uint) (h2 Selected) {
	h2 = Selected{}
	path := client.Path(api.TaskGroupBucketContentRoute).
		Inject(client.Params{
			api.ID:       id,
			api.Wildcard: "",
		})
	h2.Bucket = bucket.NewContent(h.client, path)
	return
}

// Selected task group API.
type Selected struct {
	Bucket bucket.Content
}
