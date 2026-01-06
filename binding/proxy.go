package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Proxy API.
type Proxy struct {
	client *Client
}

// Create a Proxy.
func (h *Proxy) Create(r *api2.Proxy) (err error) {
	err = h.client.Post(api2.ProxiesRoute, &r)
	return
}

// Get a Proxy by ID.
func (h *Proxy) Get(id uint) (r *api2.Proxy, err error) {
	r = &api2.Proxy{}
	path := Path(api2.ProxyRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Proxies.
func (h *Proxy) List() (list []api2.Proxy, err error) {
	list = []api2.Proxy{}
	err = h.client.Get(api2.ProxiesRoute, &list)
	return
}

// Update a Proxy.
func (h *Proxy) Update(r *api2.Proxy) (err error) {
	path := Path(api2.ProxyRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Proxy.
func (h *Proxy) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.ProxyRoute).Inject(Params{api2.ID: id}))
	return
}

// Find by Kind.
// Returns nil when not found.
func (h *Proxy) Find(kind string) (r *api2.Proxy, err error) {
	list, err := h.List()
	if err != nil {
		return
	}
	for i := range list {
		p := &list[i]
		if p.Kind == kind {
			r = p
			break
		}
	}
	return
}
