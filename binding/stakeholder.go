package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Stakeholder API.
type Stakeholder struct {
	// hub API client.
	Client *Client
}

//
// Create a Stakeholder.
func (h *Stakeholder) Create(r *api.Stakeholder) (err error) {
	err = h.Client.Post(api.StakeholdersRoot, &r)
	return
}

//
// Get a Stakeholder by ID.
func (h *Stakeholder) Get(id uint) (r *api.Stakeholder, err error) {
	r = &api.Stakeholder{}
	path := Path(api.StakeholderRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List Stakeholders.
func (h *Stakeholder) List() (list []api.Stakeholder, err error) {
	list = []api.Stakeholder{}
	err = h.Client.Get(api.StakeholdersRoot, &list)
	return
}

//
// Update a Stakeholder.
func (h *Stakeholder) Update(r *api.Stakeholder) (err error) {
	path := Path(api.StakeholderRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete a Stakeholder.
func (h *Stakeholder) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.StakeholderRoot).Inject(Params{api.ID: id}))
	return
}
