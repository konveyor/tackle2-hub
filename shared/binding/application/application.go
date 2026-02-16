package application

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) (h Application) {
	h = Application{client: client}
	return
}

// Application API.
type Application struct {
	client client.RestClient
}

// Create an Application.
func (h Application) Create(r *api.Application) (err error) {
	err = h.client.Post(api.ApplicationsRoute, r)
	return
}

// Get an Application by ID.
func (h Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := client.Path(api.ApplicationRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Applications.
func (h Application) List() (list []api.Application, err error) {
	list = []api.Application{}
	err = h.client.Get(api.ApplicationsRoute, &list)
	return
}

// Update an Application.
func (h Application) Update(r *api.Application) (err error) {
	path := client.Path(api.ApplicationRoute).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete an Application.
func (h Application) Delete(id uint) (err error) {
	path := client.
		Path(api.ApplicationRoute).
		Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// DeleteList deletes multiple applications by ID.
func (h Application) DeleteList(ids []uint) (err error) {
	err = h.client.DeleteWith(api.ApplicationsRoute, ids)
	return
}

// UpdateStakeholders updates the stakeholders for an application.
func (h Application) UpdateStakeholders(id uint, stakeholders *api.Stakeholders) (err error) {
	path := client.Path(api.AppStakeholdersRoute).Inject(client.Params{api.ID: id})
	err = h.client.Put(path, stakeholders)
	return
}

// Select returns the API for a selected application.
func (h Application) Select(id uint) (h2 Selected) {
	h2 = Selected{}
	bucketRoot := client.Path(api.AppBucketContentRoute).
		Inject(client.Params{
			api.Wildcard: "",
			api.ID:       id,
		})
	h2.Bucket = bucket.NewContent(h.client, bucketRoot)
	h2.Identity = Identity{client: h.client, appId: id}
	h2.Assessment = Assessment{client: h.client, appId: id}
	h2.Analysis = Analysis{client: h.client, appId: id}
	h2.Manifest = Manifest{client: h.client, appId: id}
	h2.Tag = Tag{client: h.client, appId: id}
	h2.Fact = Fact{client: h.client, appId: id}
	return
}

// Selected  application API.
type Selected struct {
	Bucket     bucket.Content
	Identity   Identity
	Assessment Assessment
	Analysis   Analysis
	Manifest   Manifest
	Tag        Tag
	Fact       Fact
}
