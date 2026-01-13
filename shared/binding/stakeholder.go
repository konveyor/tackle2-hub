package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Stakeholder API.
type Stakeholder struct {
	client *Client
}

// Create a Stakeholder.
func (h *Stakeholder) Create(r *api2.Stakeholder) (err error) {
	err = h.client.Post(api2.StakeholdersRoute, &r)
	return
}

// Get a Stakeholder by ID.
func (h *Stakeholder) Get(id uint) (r *api2.Stakeholder, err error) {
	r = &api2.Stakeholder{}
	path := Path(api2.StakeholderRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Stakeholders.
func (h *Stakeholder) List() (list []api2.Stakeholder, err error) {
	list = []api2.Stakeholder{}
	err = h.client.Get(api2.StakeholdersRoute, &list)
	return
}

// Update a Stakeholder.
func (h *Stakeholder) Update(r *api2.Stakeholder) (err error) {
	path := Path(api2.StakeholderRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Stakeholder.
func (h *Stakeholder) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.StakeholderRoute).Inject(Params{api2.ID: id}))
	return
}
