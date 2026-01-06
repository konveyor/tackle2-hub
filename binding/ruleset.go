package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// RuleSet API.
type RuleSet struct {
	client *Client
}

// Create a RuleSet.
func (h *RuleSet) Create(r *api2.RuleSet) (err error) {
	err = h.client.Post(api2.RuleSetsRoute, &r)
	return
}

// Get a RuleSet by ID.
func (h *RuleSet) Get(id uint) (r *api2.RuleSet, err error) {
	r = &api2.RuleSet{}
	path := Path(api2.RuleSetRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List RuleSets.
func (h *RuleSet) List() (list []api2.RuleSet, err error) {
	list = []api2.RuleSet{}
	err = h.client.Get(api2.RuleSetsRoute, &list)
	return
}

// Find RuleSets with filter.
func (h *RuleSet) Find(filter Filter) (list []api2.RuleSet, err error) {
	list = []api2.RuleSet{}
	err = h.client.Get(api2.RuleSetsRoute, &list, filter.Param())
	return
}

// Update a RuleSet.
func (h *RuleSet) Update(r *api2.RuleSet) (err error) {
	path := Path(api2.RuleSetRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a RuleSet.
func (h *RuleSet) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.RuleSetRoute).Inject(Params{api2.ID: id}))
	return
}
