package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Application API.
type Application struct {
	// hub API client.
	Client *Client
}

//
// Create an Application.
func (h *Application) Create(r *api.Application) (err error) {
	err = h.Client.Post(api.ApplicationsRoot, &r)
	return
}

//
// Get an Application by ID.
func (h *Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := Path(api.ApplicationRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List Applications.
func (h *Application) List() (list []api.Application, err error) {
	list = []api.Application{}
	err = h.Client.Get(api.ApplicationsRoot, &list)
	return
}

//
// Update an Application.
func (h *Application) Update(r *api.Application) (err error) {
	path := Path(api.ApplicationRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete an Application.
func (h *Application) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.ApplicationRoot).Inject(Params{api.ID: id}))
	return
}

//
// Bucket returns the bucket API.
func (h *Application) Bucket(id uint) (b *Bucket) {
	params := Params{
		api.Wildcard: "",
		api.ID:       id,
	}
	path := Path(api.AppBucketContentRoot).Inject(params)
	b = &Bucket{
		path:   path,
		Client: h.Client,
	}
	return
}
