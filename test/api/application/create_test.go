package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationCreate(t *testing.T) {
	samples := Samples()
	// Create on array of Applications calls subtest
	for _, application := range samples {
		t.Run(fmt.Sprintf("Create application %s", application.Name), func(t *testing.T) {

			err = Client.Post(api.ApplicationsRoot, &application)
			if err != nil {
				t.Errorf("Create error: %v", err.Error()) // Error for standard test failure or failed assertion
			}

			// The Get test not included here, but in get_test.go

			// Clean the app
			Delete(t, application)
		})
	}
}

func TestApplicationNotCreateDuplicates(t *testing.T) {
	application := Samples()[0]

	// Create sample Application.
	err = Client.Post(api.ApplicationsRoot, &application)
	if err != nil {
		t.Errorf("Create error: %v", err.Error())
	}

	// Prepare Application with duplicate Name.
	dupApplication := &api.Application{
		Name: application.Name,
	}

	// Try create the duplicate Application.
	err = Client.Post(api.ApplicationsRoot, &dupApplication)
	if err == nil {
		t.Errorf("Created duplicate application: %v", dupApplication)

		// Clean the app
		Delete(t, dupApplication)
	}

	// Clean the application.
	Delete(t, application)
}

func TestApplicationNotCreateWithoutName(t *testing.T) {
	// Prepare Application without Name.
	emptyApplication := &api.Application{
		Name: "",
	}

	// Try create the duplicate Application.
	err = Client.Post(api.ApplicationsRoot, &emptyApplication)
	if err == nil {
		t.Errorf("Created empty application: %v", emptyApplication)

		// Clean the application.
		Delete(t, emptyApplication)
	}
}
