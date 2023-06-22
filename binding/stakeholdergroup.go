package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// StakeholderGroup API.
type StakeholderGroup struct {
	// hub API client.
	Client *Client
}

//
// Create a StakeholderGroup.
func (h *StakeholderGroup) Create(r *api.StakeholderGroup) (err error) {
	err = h.Client.Post(api.StakeholderGroupsRoot, &r)
	return
}

//
// Get a StakeholderGroup by ID.
func (h *StakeholderGroup) Get(id uint) (r *api.StakeholderGroup, err error) {
	r = &api.StakeholderGroup{}
	path := Path(api.StakeholderGroupRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List StakeholderGroups.
func (h *StakeholderGroup) List() (list []api.StakeholderGroup, err error) {
	list = []api.StakeholderGroup{}
	err = h.Client.Get(api.StakeholderGroupsRoot, &list)
	return
}

//
// Update a StakeholderGroup.
func (h *StakeholderGroup) Update(r *api.StakeholderGroup) (err error) {
	path := Path(api.StakeholderGroupRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete a StakeholderGroup.
func (h *StakeholderGroup) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.StakeholderGroupRoot).Inject(Params{api.ID: id}))
	return
}
