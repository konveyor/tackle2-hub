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

//
// FindIdentity by kind.
func (h *Application) FindIdentity(id uint, kind string) (r *api.Identity, found bool, err error) {
	list := []api.Identity{}
	path := Params{api.ID: id}.inject(api.AppIdentitiesRoot)
	err = h.client.Get(path, &list)
	if err != nil {
		return
	}
	for i := range list {
		r = &list[i]
		if r.Kind == kind {
			m := r.Model()
			err = m.Decrypt(Addon.secret.Hub.Encryption.Passphrase)
			r.With(m)
			if err != nil {
				return
			}
			found = true
			break
		}
	}
	return
}
