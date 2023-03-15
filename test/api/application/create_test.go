package application

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/testclient"
)

func TestApplicationCreate(t *testing.T) {
	tests := []testclient.TestCase{
		{
			Name: "Create sample Pathfinder application",
			Application: &api.Application{
				Name:        "Pathfinder",
				Description: "Tackle Pathfinder application.",
				Repository: &api.Repository{
					Kind:   "git",
					URL:    "https://github.com/konveyor/tackle-pathfinder.git",
					Branch: "1.2.0",
				},
			},
			ShouldError: false,
		},
		{
			Name: "Create minimalist application",
			Application: &api.Application{
				Name: "App1",
			},
			ShouldError: false,
		},
		//		{
		//			Name: "Not Create application without name",
		//			Application: &api.Application{
		//				Name: "",
		//			},
		//			ShouldError: true,
		//		},
	}

	// Setup Hub API client
	hub, err := testclient.NewHubClient()
	if err != nil {
		t.Fatalf("Unable connect to Hub API: %v", err.Error())
	}

	// Execute test steps
	for _, tc := range tests {
		t.Log(tc.Name)

		// Create the application
		//err = hub.Create(&tc.Subject)
		err = hub.Post(api.ApplicationsRoot, &tc.Application)
		if err != nil && !tc.ShouldError {
			t.Errorf("Unexpected application create error: %v", err.Error())
		}
		if err == nil && tc.ShouldError {
			t.Errorf("Expected application create error and didn't get it")
		}
		t.Log(tc.Application)

		// Get the application
		var testApplication *api.Application
		err = hub.Get(fmt.Sprintf("%s/%d", api.ApplicationsRoot, tc.Application.ID), &testApplication)
		if err != nil && !tc.ShouldError {
			t.Errorf("Error getting application: %v", err.Error())
		} else {
			// Assert the application
			if !reflect.DeepEqual(tc.Application, testApplication) {
				t.Errorf("Got different application than expected: %v\n%v", tc.Application, testApplication)
			}
		}

		// Clean the application
		if err == nil && !tc.ShouldError {
			err = hub.Delete(fmt.Sprintf("%s/%d", api.ApplicationsRoot, tc.Application.ID))
			if err != nil && !tc.ShouldError {
				t.Errorf("Unexpected application delete error: %v", err.Error())
			}
		}
	}
}
