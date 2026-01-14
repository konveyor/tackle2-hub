package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Stakeholder API.
type Stakeholder struct {
	client *Client
}

// Create a Stakeholder.
func (h *Stakeholder) Create(r *api.Stakeholder) (err error) {
	err = h.client.Post(api.StakeholdersRoute, &r)
	return
}

// Get a Stakeholder by ID.
func (h *Stakeholder) Get(id uint) (r *api.Stakeholder, err error) {
	r = &api.Stakeholder{}
	path := Path(api.StakeholderRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Stakeholders.
func (h *Stakeholder) List() (list []api.Stakeholder, err error) {
	list = []api.Stakeholder{}
	err = h.client.Get(api.StakeholdersRoute, &list)
	return
}

// Update a Stakeholder.
func (h *Stakeholder) Update(r *api.Stakeholder) (err error) {
	path := Path(api.StakeholderRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Stakeholder.
func (h *Stakeholder) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.StakeholderRoute).Inject(Params{api.ID: id}))
	return
}
