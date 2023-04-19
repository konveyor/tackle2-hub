package businessservice

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Set of valid resources for tests and reuse.
func Samples() (samples map[string]api.BusinessService) {
	samples = map[string]api.BusinessService{
		"Marketing": {
			Name:        "Marketing",
			Description: "Marketing dept service.",
		},
		"Sales": {
			Name:        "Sales",
			Description: "Sales support service.",
		},
	}
	return
}

// Create a BusinessService.
func Create(r *api.BusinessService) (err error) {
	err = Client.Post(api.BusinessServicesRoot, &r)
	return
}

// Retrieve the BusinessService.
func Get(r *api.BusinessService) (err error) {
	err = Client.Get(client.Path(api.BusinessServiceRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the BusinessService.
func Update(r *api.BusinessService) (err error) {
	err = Client.Put(client.Path(api.BusinessServiceRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the BusinessService.
func Delete(r *api.BusinessService) (err error) {
	err = Client.Delete(client.Path(api.BusinessServiceRoot, client.Params{api.ID: r.ID}))
	return
}

// List BusinessServices.
func List(r []*api.BusinessService) (err error) {
	err = Client.Get(api.BusinessServicesRoot, &r)
	return
}
