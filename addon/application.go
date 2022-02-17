package addon

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Application API.
type Application struct {
	// hub API client.
	client *Client
}

//
// Get an application by ID.
func (h *Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := Params{api.ID: id}.inject(api.ApplicationRoot)
	err = h.client.Get(path, r)
	return
}

//
// List applications.
func (h *Application) List() (list []api.Application, err error) {
	list = []api.Application{}
	err = h.client.Get(api.ApplicationsRoot, &list)
	return
}

//
// Update an application by ID.
func (h *Application) Update(r *api.Application) (err error) {
	path := Params{api.ID: r.ID}.inject(api.ApplicationRoot)
	err = h.client.Put(path, r)
	if err == nil {
		Log.Info(
			"Addon updated: application.",
			"id",
			r.ID)
	}
	return
}
