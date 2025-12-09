package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// AnalysisProfile API.
type AnalysisProfile struct {
	client *Client
}

// Create a AnalysisProfile.
func (h *AnalysisProfile) Create(r *api.AnalysisProfile) (err error) {
	err = h.client.Post(api.AnalysisProfilesRoot, &r)
	return
}

// Get a AnalysisProfile by ID.
func (h *AnalysisProfile) Get(id uint) (r *api.AnalysisProfile, err error) {
	r = &api.AnalysisProfile{}
	path := Path(api.AnalysisProfileRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List AnalysisProfiles.
func (h *AnalysisProfile) List() (list []api.AnalysisProfile, err error) {
	list = []api.AnalysisProfile{}
	err = h.client.Get(api.AnalysisProfilesRoot, &list)
	return
}

// Update a AnalysisProfile.
func (h *AnalysisProfile) Update(r *api.AnalysisProfile) (err error) {
	path := Path(api.AnalysisProfileRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a AnalysisProfile.
func (h *AnalysisProfile) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.AnalysisProfileRoot).Inject(Params{api.ID: id}))
	return
}
