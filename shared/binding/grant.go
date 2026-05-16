package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Grant API.
type Grant struct {
	client RestClient
}

// Get a Grant by ID.
func (h Grant) Get(id uint) (r *api.Grant, err error) {
	r = &api.Grant{}
	path := Path(api.AuthGrantRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Grants.
func (h Grant) List() (list []api.Grant, err error) {
	list = []api.Grant{}
	err = h.client.Get(api.AuthGrantsRoute, &list)
	return
}

// Delete a Grant.
func (h Grant) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.AuthGrantRoute).Inject(Params{api.ID: id}))
	return
}
