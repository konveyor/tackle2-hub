package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// RuleSet API.
type RuleSet struct {
	// hub API client.
	client *Client
}

//
// Setting client from outside packages, temporary compatibility workaroud for addon pkg.
func (h *RuleSet) SetClient(client *Client) {
	h.client = client
}

//
// Create a RuleSet.
func (h *RuleSet) Create(r *api.RuleSet) (err error) {
	err = h.client.Post(api.RuleSetsRoot, &r)
	return
}

//
// Get a RuleSet by ID.
func (h *RuleSet) Get(id uint) (r *api.RuleSet, err error) {
	r = &api.RuleSet{}
	path := Path(api.RuleSetRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List RuleSets.
func (h *RuleSet) List() (list []api.RuleSet, err error) {
	list = []api.RuleSet{}
	err = h.client.Get(api.RuleSetsRoot, &list)
	return
}

//
// Update a RuleSet.
func (h *RuleSet) Update(r *api.RuleSet) (err error) {
	path := Path(api.RuleSetRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a RuleSet.
func (h *RuleSet) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.RuleSetRoot).Inject(Params{api.ID: id}))
	return
}