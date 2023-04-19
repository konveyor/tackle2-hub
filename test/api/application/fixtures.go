package application

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Set of valid Application resources for tests and reuse.
// Invalid application for negative tests are expected to be defined within the test methods, not here.
func Samples() (samples map[string]api.Application) {
	samples = map[string]api.Application{
		"Minimal": {
			Name: "Minimal application",
		},
		"PathfinderGit": {
			Name:        "Pathfinder",
			Description: "Tackle Pathfinder application.",
			Repository: &api.Repository{
				Kind:   "git",
				URL:    "https://github.com/konveyor/tackle-pathfinder.git",
				Branch: "1.2.0",
			},
		},

	}
	return
}

// Setup Hub API client
var Client = client.Client

// Create a Application.
func Create(r *api.Application) (err error) {
	err = Client.Post(api.ApplicationsRoot, &r)
	return
}

// Retrieve the Application.
func Get(r *api.Application) (err error) {
	err = Client.Get(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the Application.
func Update(r *api.Application) (err error) {
	err = Client.Put(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the Application.
func Delete(r *api.Application) (err error) {
	err = Client.Delete(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}))
	return
}

// List Applications.
func List(r []*api.Application) (err error) {
	err = Client.Get(api.ApplicationsRoot, &r)
	return
}
