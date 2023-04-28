package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationCreateGetDelete(t *testing.T) {
	// Create on array of Applications calls subtest
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			assert.Should(t, Application.Create(&r))

			// Try get.
			got := api.Application{}
			got.ID = r.ID
			assert.Should(t, Application.Get(&got))

			// Assert the get response.
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Try list.
			gotList := []*api.Application{}
			assert.Should(t, Application.List(gotList))

			// Assert the list response.
			foundR := api.Application{}
			for _, listR := range gotList {
				if listR.Name == r.Name && listR.ID == r.ID {
					foundR = *listR
					break
				}
			}
			if assert.FlatEqual(foundR, r) {
				t.Errorf("Different list entry error. Got %v, expected %v", foundR, r)
			}

			// Try delete.
			assert.Should(t, Application.Delete(&got))

			// Check the created application was deleted.
			err := Application.Get(&r)
			if err == nil {
				t.Fatalf("Exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestApplicationNotCreateDuplicates(t *testing.T) {
	r := Minimal

	// Create sample.
	assert.Should(t, Application.Create(&r))

	// Prepare Application with duplicate Name.
	dup := &api.Application{
		Name: r.Name,
	}

	// Try create the duplicate.
	err := Application.Create(dup)
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
	err := Application.Create(r)
	if err == nil {
		t.Errorf("Created empty application: %v", r)

		// Clean.
		assert.Must(t, Application.Delete(r))
	}
}
