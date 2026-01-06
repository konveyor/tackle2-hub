package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Platform API.
type Platform struct {
	client *Client
}

// Create a Platform.
func (h *Platform) Create(r *api2.Platform) (err error) {
	err = h.client.Post(api2.PlatformsRoute, &r)
	return
}

// Get a Platform by ID.
func (h *Platform) Get(id uint) (r *api2.Platform, err error) {
	r = &api2.Platform{}
	path := Path(api2.PlatformRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Platforms.
func (h *Platform) List() (list []api2.Platform, err error) {
	list = []api2.Platform{}
	err = h.client.Get(api2.PlatformsRoute, &list)
	return
}

// Update a Platform.
func (h *Platform) Update(r *api2.Platform) (err error) {
	path := Path(api2.PlatformRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Platform.
func (h *Platform) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.PlatformRoute).Inject(Params{api2.ID: id}))
	return
}
