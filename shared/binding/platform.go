package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Platform API.
type Platform struct {
	client *Client
}

// Create a Platform.
func (h *Platform) Create(r *api.Platform) (err error) {
	err = h.client.Post(api.PlatformsRoute, &r)
	return
}

// Get a Platform by ID.
func (h *Platform) Get(id uint) (r *api.Platform, err error) {
	r = &api.Platform{}
	path := Path(api.PlatformRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Platforms.
func (h *Platform) List() (list []api.Platform, err error) {
	list = []api.Platform{}
	err = h.client.Get(api.PlatformsRoute, &list)
	return
}

// Update a Platform.
func (h *Platform) Update(r *api.Platform) (err error) {
	path := Path(api.PlatformRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Platform.
func (h *Platform) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.PlatformRoute).Inject(Params{api.ID: id}))
	return
}
