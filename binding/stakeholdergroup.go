package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// StakeholderGroup API.
type StakeholderGroup struct {
	client *Client
}

// Create a StakeholderGroup.
func (h *StakeholderGroup) Create(r *api2.StakeholderGroup) (err error) {
	err = h.client.Post(api2.StakeholderGroupsRoute, &r)
	return
}

// Get a StakeholderGroup by ID.
func (h *StakeholderGroup) Get(id uint) (r *api2.StakeholderGroup, err error) {
	r = &api2.StakeholderGroup{}
	path := Path(api2.StakeholderGroupRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List StakeholderGroups.
func (h *StakeholderGroup) List() (list []api2.StakeholderGroup, err error) {
	list = []api2.StakeholderGroup{}
	err = h.client.Get(api2.StakeholderGroupsRoute, &list)
	return
}

// Update a StakeholderGroup.
func (h *StakeholderGroup) Update(r *api2.StakeholderGroup) (err error) {
	path := Path(api2.StakeholderGroupRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a StakeholderGroup.
func (h *StakeholderGroup) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.StakeholderGroupRoute).Inject(Params{api2.ID: id}))
	return
}
