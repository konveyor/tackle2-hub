package binding

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
// Create a Proxy.
func (h *Proxy) Create(r *api.Proxy) (err error) {
	err = h.client.Post(api.ProxiesRoot, &r)
	return
}

//
// Get a Proxy by ID.
func (h *Proxy) Get(id uint) (r *api.Proxy, err error) {
	r = &api.Proxy{}
	path := Path(api.ProxyRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Proxies.
func (h *Proxy) List() (list []api.Proxy, err error) {
	list = []api.Proxy{}
	err = h.client.Get(api.ProxiesRoot, &list)
	return
}

//
// Update a Proxy.
func (h *Proxy) Update(r *api.Proxy) (err error) {
	path := Path(api.ProxyRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a Proxy.
func (h *Proxy) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ProxyRoot).Inject(Params{api.ID: id}))
	return
}
