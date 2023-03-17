package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationCreate(t *testing.T) {
	//
	// vykašlat se na shoulderror, hodit aplikace do fixtures a tady mit jen nachsytánní dat z fixtures a skutečný test
	//
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
			EnsureDelete(t, application)
		})
	}
}

func TestApplicationNotCreateDuplicates(t *testing.T) {
	// Create sample Application.
	err = Client.Post(api.ApplicationsRoot, &Sample)
	if err != nil {
		t.Errorf("Create error: %v", err.Error())
	}

	// Prepare Application with duplicate Name.
	dupApplication := &api.Application{
		Name: Sample.Name,
	}

	// Try create the duplicate Application.
	err = Client.Post(api.ApplicationsRoot, &dupApplication)
	if err == nil {
		t.Errorf("Created duplicate application: %v", dupApplication)

		// Clean the app
		EnsureDelete(t, dupApplication)
	}

	// Clean the application.
	EnsureDelete(t, Sample)
}

func TestApplicationNotCreateWithoutName(t *testing.T) {
	// Prepare Application without Name.
	emptyApplication := &api.Application{
		Name: "",
	}

	// Try create the duplicate Application.
	err = Client.Post(api.ApplicationsRoot, &emptyApplication)
	if err == nil {
		t.Errorf("Created duplicate application: %v", emptyApplication)

		// Clean the application.
		EnsureDelete(t, emptyApplication)
	}
}
