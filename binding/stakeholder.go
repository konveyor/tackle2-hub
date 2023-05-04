package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Stakeholder API.
type Stakeholder struct {
	// hub API client.
	client *Client
}

//
// Create an Stakeholder.
func (h *Stakeholder) Create(r *api.Stakeholder) (err error) {
	err = h.client.Post(api.StakeholdersRoot, &r)
	return
}

//
// Get an Stakeholder by ID.
func (h *Stakeholder) Get(id uint) (r *api.Stakeholder, err error) {
	r = &api.Stakeholder{}
	path := Path(api.StakeholderRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Stakeholders.
func (h *Stakeholder) List() (list []api.Stakeholder, err error) {
	list = []api.Stakeholder{}
	err = h.client.Get(api.StakeholdersRoot, &list)
	return
}

//
// Update an Stakeholder.
func (h *Stakeholder) Update(r *api.Stakeholder) (err error) {
	path := Path(api.StakeholderRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete an Stakeholder.
func (h *Stakeholder) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.StakeholderRoot).Inject(Params{api.ID: id}))
	return
}
