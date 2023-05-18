package addon

import "github.com/konveyor/tackle2-hub/api"

//
// RuleSet API.
type RuleSet struct {
	// hub API client.
	client *Client
}

//
// Get a ruleset by ID.
func (h *RuleSet) Get(id uint) (r *api.RuleSet, err error) {
	r = &api.RuleSet{}
	path := Path(api.RuleSetRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List rulesets.
func (h *RuleSet) List() (list []api.RuleSet, err error) {
	list = []api.RuleSet{}
	err = h.client.Get(api.RuleSetsRoot, &list)
	if err != nil {
		return
	}
	return
}

//
// Update a ruleset by ID.
func (h *RuleSet) Update(r *api.RuleSet) (err error) {
	path := Path(api.RuleSetRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a ruleset.
func (h *RuleSet) Delete(id uint) (err error) {
	path := Path(api.RuleSetRoot).Inject(Params{api.ID: id})
	err = h.client.Delete(path)
	return
}
