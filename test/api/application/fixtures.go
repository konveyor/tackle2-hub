package application

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = client.Client
)

//if err != nil {
//	t.Fatalf("Unable connect to Hub API: %v", err.Error())
//}

//
// Set of valid Application resources for tests and reuse.
// Invalid application for negative tests are expected to be defined within the test methods, not here.
var Samples = []*api.Application{
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

//
// Creates a copy of SampleApplication for given test (copy is there to avoid tests inflence each other using the same object ref).
func CloneSamples() (samples []*api.Application) {
	raw, err := json.Marshal(Samples)
	if err != nil {
		fmt.Print("ERROR cloning samples")
	}
	json.Unmarshal(raw, &samples)
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
	err := Client.Delete(fmt.Sprintf("%s/%d", api.ApplicationsRoot, r.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
