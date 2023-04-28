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

// Create a Application.
func (h *Application) Create(r *api.Application) (err error) {
	err = h.client.Post(api.ApplicationsRoot, &r)
	return
}

// Retrieve the Application.
func (h *Application) Get(r *api.Application) (err error) {
	err = h.client.Get(Path(api.ApplicationRoot).Inject(Params{api.ID: r.ID}), &r)
	return
}

// Update the Application.
func (h *Application) Update(r *api.Application) (err error) {
	err = h.client.Put(Path(api.ApplicationRoot).Inject(Params{api.ID: r.ID}), &r)
	return
}

// Delete the Application.
func (h *Application) Delete(r *api.Application) (err error) {
	err = h.client.Delete(Path(api.ApplicationRoot).Inject(Params{api.ID: r.ID}))
	return
}

// List Applications.
func (h *Application) List(r []*api.Application) (err error) {
	err = h.client.Get(api.ApplicationsRoot, &r)
	return
}
