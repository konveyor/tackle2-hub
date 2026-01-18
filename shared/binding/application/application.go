package application

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client *client.Client) (h Application) {
	h = Application{client: client}
	return
}

// Application API.
type Application struct {
	client *client.Client
}

// Create an Application.
func (h Application) Create(r *api.Application) (err error) {
	err = h.client.Post(api.ApplicationsRoute, &r)
	return
}

// Get an Application by ID.
func (h Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := client.
		Path(api.ApplicationRoute).
		Inject(client.Params{api.ID: id})
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
	path := client.
		Path(api.ApplicationRoute).
		Inject(client.Params{api.ID: r.ID})
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

// Bucket returns the bucket API.
func (h Application) Bucket(id uint) (h2 bucket.BucketContent) {
	params := client.Params{
		api.Wildcard: "",
		api.ID:       id,
	}
	path := client.Path(api.AppBucketContentRoute).Inject(params)
	h2 = bucket.NewContent(h.client, path)
	return
}

// Tags returns the tags API.
func (h Application) Tags(id uint) (tg Tag) {
	tg = Tag{
		client: h.client,
		appId:  id,
	}
	return
}

// Facts returns the facts API.
func (h Application) Facts(id uint) (f Fact) {
	f = Fact{
		client: h.client,
		appId:  id,
	}
	return
}

// Analysis returns the analysis API.
func (h Application) Analysis(id uint) (a Analysis) {
	a = Analysis{
		client: h.client,
		appId:  id,
	}
	return
}

// Manifest returns the manifest API.
func (h Application) Manifest(id uint) (f Manifest) {
	f = Manifest{
		client: h.client,
		appId:  id,
	}
	return
}

// Identity returns the identity API.
func (h Application) Identity(id uint) (f Identity) {
	f = Identity{
		client: h.client,
		appId:  id,
	}
	return
}

// Assessment returns the assessment API.
func (h Application) Assessment(id uint) (f Assessment) {
	f = Assessment{
		client: h.client,
		appId:  id,
	}
	return
}
