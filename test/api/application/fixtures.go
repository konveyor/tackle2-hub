package application

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Hub REST API/addon Client
// Setup Hub API client
var Client, err = client.NewHubClient()

//if err != nil {
//	t.Fatalf("Unable connect to Hub API: %v", err.Error())
//}

//
// Set of valid Application resources for tests and reuse.
// Invalid application for negative tests are expected to be defined within the test methods, not here.
var SampleApplications = []*api.Application{
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
func Samples() (applications []*api.Application) {
	raw, err := json.Marshal(SampleApplications)
	if err != nil {
		fmt.Print("ERROR cloning samples")
	}
	json.Unmarshal(raw, &applications)
	return
}

//type TestCase struct {
//	Test        testclient.TestCase
//	Application *api.Application
//}

//
// Create an Application (and stop tests on failure).
func Create(t *testing.T, application *api.Application) {
	err = Client.Post(api.ApplicationsRoot, &application)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error()) // Fatal here, Error for standard test failure or failed assertion.
	}
}

//
// Delete the Application (and stop tests on failure).
func Delete(t *testing.T, application *api.Application) {
	err = Client.Delete(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
