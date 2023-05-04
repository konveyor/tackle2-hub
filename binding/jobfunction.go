package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// JobFunction API.
type JobFunction struct {
	// hub API client.
	client *Client
}

//
// Create an JobFunction.
func (h *JobFunction) Create(r *api.JobFunction) (err error) {
	err = h.client.Post(api.JobFunctionsRoot, &r)
	return
}

//
// Get an JobFunction by ID.
func (h *JobFunction) Get(id uint) (r *api.JobFunction, err error) {
	r = &api.JobFunction{}
	path := Path(api.JobFunctionRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List JobFunctions.
func (h *JobFunction) List() (list []api.JobFunction, err error) {
	list = []api.JobFunction{}
	err = h.client.Get(api.JobFunctionsRoot, &list)
	return
}

//
// Update an JobFunction.
func (h *JobFunction) Update(r *api.JobFunction) (err error) {
	path := Path(api.JobFunctionRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete an JobFunction.
func (h *JobFunction) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.JobFunctionRoot).Inject(Params{api.ID: id}))
	return
}
