package addon

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Volume API.
type Volume struct {
	// hub API client.
	client *Client
}

//
// Get a volume by ID.
func (h *Volume) Get(id uint) (r *api.Volume, err error) {
	r = &api.Volume{}
	path := Params{api.ID: id}.inject(api.VolumeRoot)
	err = h.client.Get(path, r)
	return
}

//
// List proxies.
func (h *Volume) List() (list []api.Volume, err error) {
	list = []api.Volume{}
	err = h.client.Get(api.VolumesRoot, &list)
	if err != nil {
		return
	}
	return
}

//
// Find by name.
// Returns nil when not found.
func (h *Volume) Find(name string) (r *api.Volume, err error) {
	list, err := h.List()
	if err != nil {
		return
	}
	for i := range list {
		v := &list[i]
		if v.Name == name {
			r = v
			break
		}
	}
	return
}

//
// Update a volume by ID.
func (h *Volume) Update(r *api.Volume) (err error) {
	path := Params{api.ID: r.ID}.inject(api.VolumeRoot)
	err = h.client.Put(path, r)
	return
}
