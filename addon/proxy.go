package addon

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Proxy API.
type Proxy struct {
	// hub API client.
	client *Client
}

//
// Get a proxy by ID.
func (h *Proxy) Get(id uint) (r *api.Proxy, err error) {
	r = &api.Proxy{}
	path := Params{api.ID: id}.inject(api.ProxyRoot)
	err = h.client.Get(path, r)
	return
}

//
// List proxies.
func (h *Proxy) List() (list []api.Proxy, err error) {
	list = []api.Proxy{}
	err = h.client.Get(api.ProxiesRoot, &list)
	if err != nil {
		return
	}
	return
}

//
// Update a proxy by ID.
func (h *Proxy) Update(r *api.Proxy) (err error) {
	path := Params{api.ID: r.ID}.inject(api.ProxyRoot)
	err = h.client.Put(path, r)
	return
}
