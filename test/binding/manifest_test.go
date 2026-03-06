package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
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
	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

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

	t.Cleanup(func() {
		_ = client.Manifest.Delete(manifest.ID)
	})

	// GET: List manifests
	list, err := client.Manifest.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(manifest, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the manifest and verify it matches
	retrieved, err := client.Manifest.Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(manifest, retrieved)
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

	// GET: Retrieve decrypted again and verify updates
	updated, err := client.Manifest.Decrypted().Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(manifest, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve with Decrypt and Inject - verify injected content
	expectedInjectedContent := api.Map{
		"name": "Updated Test Manifest",
		"key":  "UPDATEDKEY456",
		"database": api.Map{
			"url":      "db.updated.com",
			"user":     "updateduser",
			"password": "updatedpass",
		},
		"description": "Updated manifest using updateduser",
	}
	expectedSecret := api.Map{
		"key":      "UPDATEDKEY456",
		"user":     "updateduser",
		"password": "updatedpass",
	}

	updated, err = client.Manifest.Decrypted().Injected().Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(expectedInjectedContent, updated.Content)
	g.Expect(eq).To(BeTrue(), report)
	eq, report = cmp.Eq(expectedSecret, updated.Secret)
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the manifest
	err = client.Manifest.Delete(manifest.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Manifest.Get(manifest.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestManifestDecryptionAndInjection tests the Decrypt() and Inject() methods
func TestManifestDecryptionAndInjection(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for the manifest to reference
	application := &api.Application{
		Name:        "Test Manifest Decrypt/Inject App",
		Description: "Application for testing manifest decryption and injection",
	}
	err := client.Application.Create(application)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

	// Define manifest with secrets
	originalSecret := api.Map{
		"key":      "SECRET123",
		"user":     "admin",
		"password": "pass456",
	}
	originalContent := api.Map{
		"name": "Test Config",
		"key":  "$(key)",
		"database": api.Map{
			"url":      "db.example.com",
			"user":     "$(user)",
			"password": "$(password)",
		},
		"description": "Connect using $(user) and $(password)",
	}

	manifest := &api.Manifest{
		Application: api.Ref{
			ID:   application.ID,
			Name: application.Name,
		},
		Content: api.Map{
			"name": "Test Config",
			"key":  "$(key)",
			"database": api.Map{
				"url":      "db.example.com",
				"user":     "$(user)",
				"password": "$(password)",
			},
			"description": "Connect using $(user) and $(password)",
		},
		Secret: api.Map{
			"key":      "SECRET123",
			"user":     "admin",
			"password": "pass456",
		},
	}

	// CREATE: Create the manifest
	err = client.Manifest.Create(manifest)
	g.Expect(err).To(BeNil())
	g.Expect(manifest.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Manifest.Delete(manifest.ID)
	})

	// Verify Secret is encrypted after create
	eq, _ := cmp.Eq(originalSecret, manifest.Secret)
	g.Expect(eq).To(BeFalse(), "Secret should be encrypted after create")

	// Verify Content is unchanged after create
	eq, report := cmp.Eq(originalContent, manifest.Content)
	g.Expect(eq).To(BeTrue(), report)

	// GET without Decrypt - verify secret is encrypted
	encrypted, err := client.Manifest.Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(encrypted).NotTo(BeNil())
	eq, _ = cmp.Eq(originalSecret, encrypted.Secret)
	g.Expect(eq).To(BeFalse(), "Secret should be encrypted")
	eq, report = cmp.Eq(originalContent, encrypted.Content)
	g.Expect(eq).To(BeTrue(), report)

	// GET with Decrypt - verify secret is decrypted
	decrypted, err := client.Manifest.Decrypted().Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(decrypted).NotTo(BeNil())
	eq, report = cmp.Eq(originalSecret, decrypted.Secret)
	g.Expect(eq).To(BeTrue(), report)
	eq, report = cmp.Eq(originalContent, decrypted.Content)
	g.Expect(eq).To(BeTrue(), report)

	// GET with Inject only - verify content has injected secrets but secret is still encrypted
	injectedOnly, err := client.Manifest.Injected().Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(injectedOnly).NotTo(BeNil())
	// Secret should still be encrypted when using Inject() without Decrypt()
	eq, _ = cmp.Eq(originalSecret, injectedOnly.Secret)
	g.Expect(eq).To(BeFalse(), "Secret should be encrypted when using Inject() without Decrypt()")
	// Content should NOT have injected values because secret is encrypted
	// (injection requires decrypted secrets to work properly)

	// GET with Decrypt and Inject - verify secret is decrypted and content has injections
	decryptedInjected, err := client.Manifest.Decrypted().Injected().Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(decryptedInjected).NotTo(BeNil())
	eq, report = cmp.Eq(originalSecret, decryptedInjected.Secret)
	g.Expect(eq).To(BeTrue(), report)

	expectedInjectedContent := api.Map{
		"name": "Test Config",
		"key":  "SECRET123",
		"database": api.Map{
			"url":      "db.example.com",
			"user":     "admin",
			"password": "pass456",
		},
		"description": "Connect using admin and pass456",
	}
	eq, report = cmp.Eq(expectedInjectedContent, decryptedInjected.Content)
	g.Expect(eq).To(BeTrue(), report)

	// GET with Inject and Decrypt (reversed order) - should produce same result
	injectedDecrypted, err := client.Manifest.Injected().Decrypted().Get(manifest.ID)
	g.Expect(err).To(BeNil())
	g.Expect(injectedDecrypted).NotTo(BeNil())
	eq, report = cmp.Eq(originalSecret, injectedDecrypted.Secret)
	g.Expect(eq).To(BeTrue(), report)
	eq, report = cmp.Eq(expectedInjectedContent, injectedDecrypted.Content)
	g.Expect(eq).To(BeTrue(), report)

	// LIST without Decrypt - verify secrets are encrypted
	encryptedList, err := client.Manifest.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(encryptedList)).To(BeNumerically(">", 0))
	found := false
	for _, m := range encryptedList {
		if m.ID == manifest.ID {
			found = true
			eq, _ = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeFalse(), "Secret should be encrypted in list")
			eq, report = cmp.Eq(originalContent, m.Content)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// LIST with Decrypt - verify secrets are decrypted
	decryptedList, err := client.Manifest.Decrypted().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(decryptedList)).To(BeNumerically(">", 0))
	found = false
	for _, m := range decryptedList {
		if m.ID == manifest.ID {
			found = true
			eq, report = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeTrue(), report)
			eq, report = cmp.Eq(originalContent, m.Content)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// LIST with Inject only - secret still encrypted
	injectedOnlyList, err := client.Manifest.Injected().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(injectedOnlyList)).To(BeNumerically(">", 0))
	found = false
	for _, m := range injectedOnlyList {
		if m.ID == manifest.ID {
			found = true
			eq, _ = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeFalse(), "Secret should be encrypted with Inject() only")
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// LIST with Decrypt and Inject - verify secrets are decrypted and content has injections
	decryptedInjectedList, err := client.Manifest.Decrypted().Injected().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(decryptedInjectedList)).To(BeNumerically(">", 0))
	found = false
	for _, m := range decryptedInjectedList {
		if m.ID == manifest.ID {
			found = true
			eq, report = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeTrue(), report)
			eq, report = cmp.Eq(expectedInjectedContent, m.Content)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// FIND without Decrypt - verify secrets are encrypted
	filter := binding.Filter{}
	filter.And("application.id").Eq(int(application.ID))
	encryptedFound, err := client.Manifest.Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(encryptedFound)).To(BeNumerically(">", 0))
	found = false
	for _, m := range encryptedFound {
		if m.ID == manifest.ID {
			found = true
			eq, _ = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeFalse(), "Secret should be encrypted in find")
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// FIND with Decrypt - verify secrets are decrypted
	decryptedFound, err := client.Manifest.Decrypted().Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(decryptedFound)).To(BeNumerically(">", 0))
	found = false
	for _, m := range decryptedFound {
		if m.ID == manifest.ID {
			found = true
			eq, report = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// FIND with Inject only - secret still encrypted
	injectedOnlyFound, err := client.Manifest.Injected().Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(injectedOnlyFound)).To(BeNumerically(">", 0))
	found = false
	for _, m := range injectedOnlyFound {
		if m.ID == manifest.ID {
			found = true
			eq, _ = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeFalse(), "Secret should be encrypted with Inject() only")
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// FIND with Decrypt and Inject - verify secrets are decrypted and content has injections
	decryptedInjectedFound, err := client.Manifest.Decrypted().Injected().Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(decryptedInjectedFound)).To(BeNumerically(">", 0))
	found = false
	for _, m := range decryptedInjectedFound {
		if m.ID == manifest.ID {
			found = true
			eq, report = cmp.Eq(originalSecret, m.Secret)
			g.Expect(eq).To(BeTrue(), report)
			eq, report = cmp.Eq(expectedInjectedContent, m.Content)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())
}
