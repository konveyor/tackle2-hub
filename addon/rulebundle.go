package addon

import "github.com/konveyor/tackle2-hub/api"

//
// RuleBundle API.
type RuleBundle struct {
	// hub API client.
	client *Client
}

//
// Get a bundle by ID.
func (h *RuleBundle) Get(id uint) (r *api.RuleBundle, err error) {
	r = &api.RuleBundle{}
	path := Params{api.ID: id}.inject(api.RuleBundleRoot)
	err = h.client.Get(path, r)
	return
}

//
// List bundles.
func (h *RuleBundle) List() (list []api.RuleBundle, err error) {
	list = []api.RuleBundle{}
	err = h.client.Get(api.RuleBundlesRoot, &list)
	if err != nil {
		return
	}
	return
}

//
// Update a bundle by ID.
func (h *RuleBundle) Update(r *api.RuleBundle) (err error) {
	path := Params{api.ID: r.ID}.inject(api.RuleBundleRoot)
	err = h.client.Put(path, r)
	return
}

//
// Delete a bundle.
func (h *RuleBundle) Delete(id uint) (err error) {
	path := Params{api.ID: id}.inject(api.RuleBundleRoot)
	err = h.client.Delete(path)
	return
}
