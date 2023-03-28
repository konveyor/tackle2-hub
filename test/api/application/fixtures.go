package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = client.Client
)

//
// Set of valid Application resources for tests and reuse.
// Invalid application for negative tests are expected to be defined within the test methods, not here.
func Samples() (samples []api.Application) {
	samples = []api.Application{
		{
			Name:        "Pathfinder",
			Description: "Tackle Pathfinder application.",
			Repository: &api.Repository{
				Kind:   "git",
				URL:    "https://github.com/konveyor/tackle-pathfinder.git",
				Branch: "1.2.0",
			},
		},
		{
			Name: "Minimal application",
		},
	}
	return
}

//
// Create an Application (and stop tests on failure).
func Create(t *testing.T, r *api.Application) {
	err := Client.Post(api.ApplicationsRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error()) // Fatal here, Error for standard test failure or failed assertion.
	}
}

//
// Delete the Application (and stop tests on failure).
func Delete(t *testing.T, r *api.Application) {
	err := Client.Delete(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
