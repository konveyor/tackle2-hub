package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// RuleSet API.
type RuleSet struct {
	client *Client
}

// Create a RuleSet.
func (h RuleSet) Create(r *api.RuleSet) (err error) {
	err = h.client.Post(api.RuleSetsRoute, r)
	return
}

// Get a RuleSet by ID.
func (h RuleSet) Get(id uint) (r *api.RuleSet, err error) {
	r = &api.RuleSet{}
	path := Path(api.RuleSetRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List RuleSets.
func (h RuleSet) List() (list []api.RuleSet, err error) {
	list = []api.RuleSet{}
	err = h.client.Get(api.RuleSetsRoute, &list)
	return
}

// Find RuleSets with filter.
func (h RuleSet) Find(filter Filter) (list []api.RuleSet, err error) {
	list = []api.RuleSet{}
	err = h.client.Get(api.RuleSetsRoute, &list, filter.Param())
	return
}

// Update a RuleSet.
func (h RuleSet) Update(r *api.RuleSet) (err error) {
	path := Path(api.RuleSetRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a RuleSet.
func (h RuleSet) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.RuleSetRoute).Inject(Params{api.ID: id}))
	return
}
