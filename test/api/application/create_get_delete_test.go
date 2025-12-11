package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationCreateGetDelete(t *testing.T) {
	// Create on array of Applications calls subtest
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Identities.
			direct := &api.Identity{
				Name: "direct",
				Kind: "Test",
			}
			err := RichClient.Identity.Create(direct)
			assert.Must(t, err)
			defer func() {
				_ = RichClient.Identity.Delete(direct.ID)
			}()
			direct2 := &api.Identity{
				Name: "direct2",
				Kind: "Other",
			}
			err = RichClient.Identity.Create(direct2)
			assert.Must(t, err)
			defer func() {
				_ = RichClient.Identity.Delete(direct2.ID)
			}()
			r.Identities = append(
				r.Identities,
				api.IdentityRef{ID: direct.ID, Role: "A"},
				api.IdentityRef{ID: direct2.ID, Role: "B"})

			// Create
			assert.Should(t, Application.Create(&r))

			// Try get.
			got, err := Application.Get(r.ID)
			assert.Should(t, err)

			// Assert the get response.
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Try list.
			gotList, err := Application.List()
			assert.Should(t, err)

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
			assert.Should(t, Application.Delete(got.ID))

			// Check the created application was deleted.
			_, err = Application.Get(r.ID)
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
		assert.Must(t, Application.Delete(dup.ID))
	}

	// Clean.
	assert.Must(t, Application.Delete(r.ID))
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
		assert.Must(t, Application.Delete(r.ID))
	}
}
