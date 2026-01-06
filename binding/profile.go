package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Analysis profile API.
type AnalysisProfile struct {
	client *Client
}

// Create a profile.
func (h *AnalysisProfile) Create(r *api2.AnalysisProfile) (err error) {
	err = h.client.Post(api2.AnalysisProfilesRoute, &r)
	return
}

// Get a profile by ID.
func (h *AnalysisProfile) Get(id uint) (r *api2.AnalysisProfile, err error) {
	r = &api2.AnalysisProfile{}
	path := Path(api2.AnalysisProfileRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// GetBundle downloads a profile bundle to the specified destination.
func (h *AnalysisProfile) GetBundle(id uint, destination string) (err error) {
	path := Path(api2.AnalysisProfileBundle).Inject(Params{api2.ID: id})
	err = h.client.FileGet(path, destination)
	if err != nil {
		return
	}
	return
}

// List profiles.
func (h *AnalysisProfile) List() (list []api2.AnalysisProfile, err error) {
	list = []api2.AnalysisProfile{}
	err = h.client.Get(api2.AnalysisProfilesRoute, &list)
	return
}

// Update a profile.
func (h *AnalysisProfile) Update(r *api2.AnalysisProfile) (err error) {
	path := Path(api2.AnalysisProfileRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a profile.
func (h *AnalysisProfile) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.AnalysisProfileRoute).Inject(Params{api2.ID: id}))
	return
}
