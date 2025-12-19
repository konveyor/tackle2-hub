package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Analysis profile API.
type AnalysisProfile struct {
	client *Client
}

// Create a profile.
func (h *AnalysisProfile) Create(r *api.AnalysisProfile) (err error) {
	err = h.client.Post(api.AnalysisProfilesRoute, &r)
	return
}

// Get a profile by ID.
func (h *AnalysisProfile) Get(id uint) (r *api.AnalysisProfile, err error) {
	r = &api.AnalysisProfile{}
	path := Path(api.AnalysisProfileRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// GetBundle downloads a profile bundle to the specified destination.
func (h *AnalysisProfile) GetBundle(id uint, destination string) (err error) {
	path := Path(api.AnalysisProfileBundle).Inject(Params{api.ID: id})
	err = h.client.FileGet(path, destination)
	if err != nil {
		return
	}
	return
}

// List profiles.
func (h *AnalysisProfile) List() (list []api.AnalysisProfile, err error) {
	list = []api.AnalysisProfile{}
	err = h.client.Get(api.AnalysisProfilesRoute, &list)
	return
}

// Update a profile.
func (h *AnalysisProfile) Update(r *api.AnalysisProfile) (err error) {
	path := Path(api.AnalysisProfileRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a profile.
func (h *AnalysisProfile) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.AnalysisProfileRoute).Inject(Params{api.ID: id}))
	return
}
