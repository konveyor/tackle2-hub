package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// StakeholderGroup API.
type StakeholderGroup struct {
	client *Client
}

// Create a StakeholderGroup.
func (h StakeholderGroup) Create(r *api.StakeholderGroup) (err error) {
	err = h.client.Post(api.StakeholderGroupsRoute, r)
	return
}

// Get a StakeholderGroup by ID.
func (h StakeholderGroup) Get(id uint) (r *api.StakeholderGroup, err error) {
	r = &api.StakeholderGroup{}
	path := Path(api.StakeholderGroupRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List StakeholderGroups.
func (h StakeholderGroup) List() (list []api.StakeholderGroup, err error) {
	list = []api.StakeholderGroup{}
	err = h.client.Get(api.StakeholderGroupsRoute, &list)
	return
}

// Update a StakeholderGroup.
func (h StakeholderGroup) Update(r *api.StakeholderGroup) (err error) {
	path := Path(api.StakeholderGroupRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a StakeholderGroup.
func (h StakeholderGroup) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.StakeholderGroupRoute).Inject(Params{api.ID: id}))
	return
}
