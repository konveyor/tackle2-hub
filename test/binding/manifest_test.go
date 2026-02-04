package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestManifest(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for the manifest to reference
	application := &api.Application{
		Name:        "Test Manifest App",
		Description: "Application for manifest testing",
	}
	err := client.Application.Create(application)
	g.Expect(err).To(BeNil())
	defer func() {
		_ = client.Application.Delete(application.ID)
	}()

	// Define the manifest to create
	manifest := &api.Manifest{
		Application: api.Ref{
			ID:   application.ID,
			Name: application.Name,
		},
		Content: api.Map{
			"name": "Test Manifest",
			"key":  "$(key)",
			"database": api.Map{
				"url":      "db.test.com",
				"user":     "$(user)",
				"password": "$(password)",
			},
			"description": "Connect using $(user) and $(password)",
		},
		Secret: api.Map{
			"key":      "TESTKEY123",
			"user":     "testuser",
			"password": "testpass",
		},
	}

	// CREATE: Create the manifest
	err = client.Manifest.Create(manifest)
	g.Expect(err).To(BeNil())
	g.Expect(manifest.ID).NotTo(BeZero())

	defer func() {
		_ = client.Manifest.Delete(manifest.ID)
	}()

	// GET: Retrieve the manifest and verify it matches
	retrieved, err := client.Manifest.Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(manifest, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the manifest
	manifest.Content = api.Map{
		"name": "Updated Test Manifest",
		"key":  "$(key)",
		"database": api.Map{
			"url":      "db.updated.com",
			"user":     "$(user)",
			"password": "$(password)",
		},
		"description": "Updated manifest using $(user)",
	}
	manifest.Secret = api.Map{
		"key":      "UPDATEDKEY456",
		"user":     "updateduser",
		"password": "updatedpass",
	}

	err = client.Manifest.Update(manifest)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Manifest.Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(manifest, updated, "UpdateUser", "Secret")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the manifest
	err = client.Manifest.Delete(manifest.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Manifest.Get(manifest.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
