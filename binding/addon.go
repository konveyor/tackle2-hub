package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Addon API.
type Addon struct {
	client *Client
}

// Get an Addon by name.
func (h *Addon) Get(name string) (r *api2.Addon, err error) {
	r = &api2.Addon{}
	path := Path(api2.AddonRoute).Inject(Params{api2.Name: name})
	err = h.client.Get(path, r)
	return
}

// List Addons.
func (h *Addon) List() (list []api2.Addon, err error) {
	list = []api2.Addon{}
	err = h.client.Get(api2.AddonsRoute, &list)
	return
}
