package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Assessment API.
type Assessment struct {
	client *Client
}

// Get a Assessment by ID.
func (h *Assessment) Get(id uint) (r *api.Assessment, err error) {
	r = &api.Assessment{}
	path := Path(api.AssessmentRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Assessments.
func (h *Assessment) List() (list []api.Assessment, err error) {
	list = []api.Assessment{}
	err = h.client.Get(api.AssessmentsRoute, &list)
	return
}

// Update a Assessment.
func (h *Assessment) Update(r *api.Assessment) (err error) {
	path := Path(api.AssessmentRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Assessment.
func (h *Assessment) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.AssessmentRoute).Inject(Params{api.ID: id}))
	return
}
