package binding

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
// Create a Application.
func (h *Application) Create(r *api.Application) (err error) {
	err = h.client.Post(api.ApplicationsRoot, &r)
	return
}

//
// Get an application by ID.
func (h *Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := Path(api.ApplicationRoot).Inject(Params{api.ID: id})
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
// Update an application.
func (h *Application) Update(r *api.Application) (err error) {
	path := Path(api.ApplicationRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete the Application.
func (h *Application) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ApplicationRoot).Inject(Params{api.ID: id}))
	return
}
