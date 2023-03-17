package application

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/testclient"
)

// Hub REST API/addon Client
// Setup Hub API client
var Client, err = testclient.NewHubClient()

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

func Samples() (applications []*api.Application) {
	raw, err := json.Marshal(SampleApplications)
	if err != nil {
		fmt.Print("ERROR cloning samples")
	}
	json.Unmarshal(raw, &applications)
	return
}

//deepcopy with json marshalling

// TODO: deepcopy on read to avoid modifying by a test before using by other tests?

//
// A single valid Application to be used as a sample for testing.
var Sample = SampleApplications[0]

func EnsureDelete(t *testing.T, application *api.Application) {
	err = Client.Delete(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID))
	if err != nil {
		t.Fatalf("Fatal error: %v", err.Error())
	}
	// Ensure would mean Fatal error in failed or ignore if failed?
}

//type TestCase struct {
//	Test        testclient.TestCase
//	Application *api.Application
//}

func Create(t *testing.T, application *api.Application) {
	err = Client.Post(api.ApplicationsRoot, &application)
	if err != nil {
		t.Fatalf("Create error: %v", err.Error()) // Error for standard test failure or failed assertion
	}
}
