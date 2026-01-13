package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Assessment API.
type Assessment struct {
	client *Client
}

// Get a Assessment by ID.
func (h *Assessment) Get(id uint) (r *api2.Assessment, err error) {
	r = &api2.Assessment{}
	path := Path(api2.AssessmentRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Assessments.
func (h *Assessment) List() (list []api2.Assessment, err error) {
	list = []api2.Assessment{}
	err = h.client.Get(api2.AssessmentsRoute, &list)
	return
}

// Update a Assessment.
func (h *Assessment) Update(r *api2.Assessment) (err error) {
	path := Path(api2.AssessmentRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Assessment.
func (h *Assessment) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.AssessmentRoute).Inject(Params{api2.ID: id}))
	return
}
