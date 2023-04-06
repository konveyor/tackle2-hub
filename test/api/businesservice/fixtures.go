package businessservice

import (
	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = c.Client
)

//
// Set of valid resources for tests and reuse.
func Samples() (samples []api.BusinessService) {
	samples = []api.BusinessService{
		{
			Name:        "Marketing",
			Description: "Marketing dept service.",
		},
		{
			Name:        "Sales",
			Description: "Sales support service.",
		},
	}
	return
}

//
// Create a BusinessService.
func Create(r *api.BusinessService) (err error) {
	err = Client.Post(api.BusinessServicesRoot, &r)
	return
}

//
// Retrieve the BusinessService.
func Get(r *api.BusinessService) (err error) {
	err = Client.Get(c.Path(api.BusinessServiceRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Update the BusinessService.
func Update(r *api.BusinessService) (err error) {
	err = Client.Put(c.Path(api.BusinessServiceRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Delete the BusinessService.
func Delete(r *api.BusinessService) (err error) {
	err = Client.Delete(c.Path(api.BusinessServiceRoot, c.Params{api.ID: r.ID}))
	return
}

//
// List BusinessServices.
func List(r []*api.BusinessService) (err error) {
	err = Client.Get(api.BusinessServicesRoot, &r)
	return
}
