package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Target API.
type Target struct {
	client *Client
}

// Create a Target.
func (h *Target) Create(r *api.Target) (err error) {
	err = h.client.Post(api.TargetsRoute, &r)
	return
}

// Get a Target by ID.
func (h *Target) Get(id uint) (r *api.Target, err error) {
	r = &api.Target{}
	path := Path(api.TargetRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Targets.
func (h *Target) List() (list []api.Target, err error) {
	list = []api.Target{}
	err = h.client.Get(api.TargetsRoute, &list)
	return
}

// Update a Target.
func (h *Target) Update(r *api.Target) (err error) {
	path := Path(api.TargetRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Target.
func (h *Target) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TargetRoute).Inject(Params{api.ID: id}))
	return
}
