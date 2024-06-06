package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Addon API.
type Addon struct {
	client *Client
}

// Get an Addon by name.
func (h *Addon) Get(name string) (r *api.Addon, err error) {
	r = &api.Addon{}
	path := Path(api.AddonRoot).Inject(Params{api.Name: name})
	err = h.client.Get(path, r)
	return
}

// List Addons.
func (h *Addon) List() (list []api.Addon, err error) {
	list = []api.Addon{}
	err = h.client.Get(api.AddonsRoot, &list)
	return
}
