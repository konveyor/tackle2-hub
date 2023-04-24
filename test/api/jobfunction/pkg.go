package jobfunction

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Create a JobFunction.
func Create(r *api.JobFunction) (err error) {
	err = Client.Post(api.JobFunctionsRoot, &r)
	return
}

// Retrieve the JobFunction.
func Get(r *api.JobFunction) (err error) {
	err = Client.Get(client.Path(api.JobFunctionRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the JobFunction.
func Update(r *api.JobFunction) (err error) {
	err = Client.Put(client.Path(api.JobFunctionRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the JobFunction.
func Delete(r *api.JobFunction) (err error) {
	err = Client.Delete(client.Path(api.JobFunctionRoot, client.Params{api.ID: r.ID}))
	return
}

// List JobFunctions.
func List(r []*api.JobFunction) (err error) {
	err = Client.Get(api.JobFunctionsRoot, &r)
	return
}
