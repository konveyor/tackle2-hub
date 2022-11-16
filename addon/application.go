package addon

import (
	"github.com/konveyor/tackle2-hub/api"
	"strconv"
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
	p1 := Param{
		Key:   api.AppId,
		Value: strconv.Itoa(int(id)),
	}
	p2 := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	path := Params{api.ID: id}.inject(api.IdentitiesRoot)
	err = h.client.Get(path, &list, p1, p2)
	if err != nil {
		return
	}
	for i := range list {
		r = &list[i]
		if r.Kind == kind {
			m := r.Model()
			r.With(m)
			found = true
			break
		}
	}
	return
}
