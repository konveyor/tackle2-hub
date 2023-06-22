package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// JobFunction API.
type JobFunction struct {
	// hub API client.
	Client *Client
}

//
// Create a JobFunction.
func (h *JobFunction) Create(r *api.JobFunction) (err error) {
	err = h.Client.Post(api.JobFunctionsRoot, &r)
	return
}

//
// Get a JobFunction by ID.
func (h *JobFunction) Get(id uint) (r *api.JobFunction, err error) {
	r = &api.JobFunction{}
	path := Path(api.JobFunctionRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List JobFunctions.
func (h *JobFunction) List() (list []api.JobFunction, err error) {
	list = []api.JobFunction{}
	err = h.Client.Get(api.JobFunctionsRoot, &list)
	return
}

//
// Update a JobFunction.
func (h *JobFunction) Update(r *api.JobFunction) (err error) {
	path := Path(api.JobFunctionRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete a JobFunction.
func (h *JobFunction) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.JobFunctionRoot).Inject(Params{api.ID: id}))
	return
}
