package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationCreateGetDelete(t *testing.T) {
	// Create on array of Applications calls subtest
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			err := Client.Post(api.ApplicationsRoot, &r)
			if err != nil {
				t.Errorf(err.Error()) // Error for standard test failure or failed assertion
			}
			rPath := client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID})

			// Try get.
			got := api.Application{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf("Get error: %v", err.Error())
			}

			// Assert the get response.
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Try list.
			gotList := []api.Application{}
			err = Client.Get(api.ApplicationsRoot, &gotList)
			if err != nil {
				t.Errorf("List error: %v", err.Error())
			}

			// Assert the list response.
			foundR := api.Application{}
			for _, listR := range gotList {
				if listR.Name == r.Name && listR.ID == r.ID {
					foundR = listR
					break
				}
			}
			if assert.FlatEqual(foundR, r) {
				t.Errorf("Different list entry error. Got %v, expected %v", foundR, r)
			}

			// Try delete.
			err = Client.Delete(rPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check the it was deleted.
			err = Client.Get(rPath, &r)
			if err == nil {
				t.Fatalf("Exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestApplicationNotCreateDuplicates(t *testing.T) {
	r := Minimal

	// Create sample.
	err := Client.Post(api.ApplicationsRoot, &r)
	if err != nil {
		t.Errorf("Create error: %v", err.Error())
	}

	// Prepare Application with duplicate Name.
	dup := &api.Application{
		Name: r.Name,
	}

	// Try create the duplicate.
	err = Client.Post(api.ApplicationsRoot, &dup)
	if err == nil {
		t.Errorf("Created duplicate application: %v", dup)

		// Clean the duplicate.
		assert.Must(t, Application.Delete(dup))
	}

	// Clean.
	assert.Must(t, Application.Delete(&r))
}

func TestApplicationNotCreateWithoutName(t *testing.T) {
	// Prepare Application without Name.
	r := &api.Application{
		Name: "",
	}

	// Try create the duplicate Application.
	err := Client.Post(api.ApplicationsRoot, &r)
	if err == nil {
		t.Errorf("Created empty application: %v", r)

		// Clean.
		assert.Must(t, Application.Delete(r))
	}
}
