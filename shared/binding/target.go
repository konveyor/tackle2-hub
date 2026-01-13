package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Target API.
type Target struct {
	client *Client
}

// Create a Target.
func (h *Target) Create(r *api2.Target) (err error) {
	err = h.client.Post(api2.TargetsRoute, &r)
	return
}

// Get a Target by ID.
func (h *Target) Get(id uint) (r *api2.Target, err error) {
	r = &api2.Target{}
	path := Path(api2.TargetRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Targets.
func (h *Target) List() (list []api2.Target, err error) {
	list = []api2.Target{}
	err = h.client.Get(api2.TargetsRoute, &list)
	return
}

// Update a Target.
func (h *Target) Update(r *api2.Target) (err error) {
	path := Path(api2.TargetRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Target.
func (h *Target) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.TargetRoute).Inject(Params{api2.ID: id}))
	return
}
