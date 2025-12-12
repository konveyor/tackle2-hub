package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// JobFunction API.
type JobFunction struct {
	client *Client
}

// Create a JobFunction.
func (h *JobFunction) Create(r *api.JobFunction) (err error) {
	err = h.client.Post(api.JobFunctionsRoute, &r)
	return
}

// Get a JobFunction by ID.
func (h *JobFunction) Get(id uint) (r *api.JobFunction, err error) {
	r = &api.JobFunction{}
	path := Path(api.JobFunctionRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List JobFunctions.
func (h *JobFunction) List() (list []api.JobFunction, err error) {
	list = []api.JobFunction{}
	err = h.client.Get(api.JobFunctionsRoute, &list)
	return
}

// Update a JobFunction.
func (h *JobFunction) Update(r *api.JobFunction) (err error) {
	path := Path(api.JobFunctionRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a JobFunction.
func (h *JobFunction) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.JobFunctionRoute).Inject(Params{api.ID: id}))
	return
}
