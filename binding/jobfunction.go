package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// JobFunction API.
type JobFunction struct {
	client *Client
}

// Create a JobFunction.
func (h *JobFunction) Create(r *api2.JobFunction) (err error) {
	err = h.client.Post(api2.JobFunctionsRoute, &r)
	return
}

// Get a JobFunction by ID.
func (h *JobFunction) Get(id uint) (r *api2.JobFunction, err error) {
	r = &api2.JobFunction{}
	path := Path(api2.JobFunctionRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List JobFunctions.
func (h *JobFunction) List() (list []api2.JobFunction, err error) {
	list = []api2.JobFunction{}
	err = h.client.Get(api2.JobFunctionsRoute, &list)
	return
}

// Update a JobFunction.
func (h *JobFunction) Update(r *api2.JobFunction) (err error) {
	path := Path(api2.JobFunctionRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a JobFunction.
func (h *JobFunction) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.JobFunctionRoute).Inject(Params{api2.ID: id}))
	return
}
