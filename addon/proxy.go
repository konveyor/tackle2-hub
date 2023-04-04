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
	path := Path(api.ProxyRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// Find by Kind.
// Returns nil when not found.
func (h *Proxy) Find(kind string) (r *api.Proxy, err error) {
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
	path := Path(api.ProxyRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}
