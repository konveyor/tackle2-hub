package binding

import (
	"errors"
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

// TestApplication tests the main Application CRUD operations
func TestApplication(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create BusinessService
	businessService := &api.BusinessService{
		Name:        "Test Business Service",
		Description: "Business service for application testing",
	}
	err := client.BusinessService.Create(businessService)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.BusinessService.Delete(businessService.ID)
	})

	// Create Owner stakeholder
	owner := &api.Stakeholder{
		Name:  "Test Owner",
		Email: "owner@test.com",
	}
	err = client.Stakeholder.Create(owner)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Stakeholder.Delete(owner.ID)
	})

	// Create Contributors
	contributor1 := &api.Stakeholder{
		Name:  "Test Contributor 1",
		Email: "contributor1@test.com",
	}
	err = client.Stakeholder.Create(contributor1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Stakeholder.Delete(contributor1.ID)
	})

	contributor2 := &api.Stakeholder{
		Name:  "Test Contributor 2",
		Email: "contributor2@test.com",
	}
	err = client.Stakeholder.Create(contributor2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Stakeholder.Delete(contributor2.ID)
	})

	// CREATE: Create a fully populated application
	app := &api.Application{
		Name:        "Test Application",
		Description: "This is a test application for CRUD operations",
		Comments:    "Initial comments",
		Binary:      "com.test:test-app:1.0.0:jar",
		Repository: &api.Repository{
			Kind:   "git",
			URL:    "https://github.com/test/test-app.git",
			Branch: "main",
			Path:   "",
		},
		BusinessService: &api.Ref{
			ID:   businessService.ID,
			Name: businessService.Name,
		},
		Owner: &api.Ref{
			ID:   owner.ID,
			Name: owner.Name,
		},
		Contributors: []api.Ref{
			{ID: contributor1.ID, Name: contributor1.Name},
			{ID: contributor2.ID, Name: contributor2.Name},
		},
	}
	err = client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// LIST: List applications and verify
	list, err := client.Application.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(BeNumerically(">", 0))
	found := false
	for _, a := range list {
		if a.ID == app.ID {
			found = true
			eq, report := cmp.Eq(app, &a)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// GET: Retrieve the application and verify it matches
	retrieved, err := client.Application.Get(app.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(app, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the application
	app.Name = "Updated Test Application"
	app.Description = "This is an updated test application"
	app.Comments = "Updated comments"
	app.Binary = "com.test:test-app:2.0.0:war"
	app.Repository = &api.Repository{
		Kind:   "git",
		URL:    "https://github.com/test/test-app-v2.git",
		Branch: "develop",
		Path:   "/src",
	}
	err = client.Application.Update(app)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Application.Get(app.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(app, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the application
	err = client.Application.Delete(app.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Application.Get(app.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestApplicationIdentity tests the Application.Select().Identity subresource
func TestApplicationIdentity(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an identity for testing
	identity := &api.Identity{
		Name:     "test-app-identity",
		Kind:     "git",
		User:     "test-user",
		Password: "test-password-123",
	}
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	g.Expect(identity.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Identity",
		Description: "Application for testing identity subresource",
	}
	err = client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// LIST: Verify initially empty
	list, err := selected.Identity.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(0))

	// Test Direct identity association (not implemented in current code, placeholder for future)
	// Direct identities would be associated via a role field

	// Test Indirect identity lookup
	// Note: Indirect searches for identities by kind and default flag
	foundIdentity, found, err := selected.Identity.Indirect(identity.Kind)
	g.Expect(err).To(BeNil())
	if found {
		g.Expect(foundIdentity).NotTo(BeNil())
		g.Expect(foundIdentity.Kind).To(Equal(identity.Kind))
	}

	// Test Search API
	search := selected.Identity.Search()
	foundIdentity, found, err = search.Indirect(identity.Kind).Find()
	g.Expect(err).To(BeNil())
	if found {
		g.Expect(foundIdentity).NotTo(BeNil())
		g.Expect(foundIdentity.Kind).To(Equal(identity.Kind))
	}
}

// TestApplicationIdentityWithRoles tests the Application.Select().Identity Direct search with roles
func TestApplicationIdentityWithRoles(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create identities for testing
	// Direct identity with role "source"
	directSource := &api.Identity{
		Name: "direct-source",
		Kind: "Test",
	}
	err := client.Identity.Create(directSource)
	g.Expect(err).To(BeNil())
	g.Expect(directSource.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Identity.Delete(directSource.ID)
	})

	// Direct identity with role "asset"
	directAsset := &api.Identity{
		Name: "direct-asset",
		Kind: "Other",
	}
	err = client.Identity.Create(directAsset)
	g.Expect(err).To(BeNil())
	g.Expect(directAsset.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Identity.Delete(directAsset.ID)
	})

	// Indirect default identity for "Other" kind
	indirectOther := &api.Identity{
		Name:    "indirect-other",
		Kind:    "Other",
		Default: true,
	}
	err = client.Identity.Create(indirectOther)
	g.Expect(err).To(BeNil())
	g.Expect(indirectOther.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Identity.Delete(indirectOther.ID)
	})

	// Indirect default identity for "Test" kind
	indirectTest := &api.Identity{
		Kind:    "Test",
		Name:    "indirect-test",
		Default: true,
	}
	err = client.Identity.Create(indirectTest)
	g.Expect(err).To(BeNil())
	g.Expect(indirectTest.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Identity.Delete(indirectTest.ID)
	})

	// CREATE: Create application with direct identities assigned with roles
	app := &api.Application{
		Name: "Test App for Identity with Roles",
		Identities: []api.IdentityRef{
			{ID: directSource.ID, Role: "source"},
			{ID: directAsset.ID, Role: "asset"},
		},
	}
	err = client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// LIST: Verify direct identities are assigned
	list, err := selected.Identity.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(2))

	// SEARCH: Find direct identity with role "source"
	foundIdentity, found, err := selected.Identity.Search().
		Direct("source").
		Indirect(indirectOther.Kind).
		Find()
	g.Expect(err).To(BeNil())
	g.Expect(found).To(BeTrue())
	g.Expect(foundIdentity).NotTo(BeNil())
	g.Expect(foundIdentity.ID).To(Equal(directSource.ID))

	// SEARCH: Find indirect identity when no direct role specified
	foundIdentity, found, err = selected.Identity.Search().
		Direct("").
		Indirect(indirectOther.Kind).
		Find()
	g.Expect(err).To(BeNil())
	g.Expect(found).To(BeTrue())
	g.Expect(foundIdentity).NotTo(BeNil())
	g.Expect(foundIdentity.ID).To(Equal(indirectOther.ID))

	// SEARCH: Find direct identity with role "asset" (multiple Direct calls)
	foundIdentity, found, err = selected.Identity.Search().
		Direct("none").
		Direct("asset").
		Indirect(indirectOther.Kind).
		Find()
	g.Expect(err).To(BeNil())
	g.Expect(found).To(BeTrue())
	g.Expect(foundIdentity).NotTo(BeNil())
	g.Expect(foundIdentity.ID).To(Equal(directAsset.ID))

	// SEARCH: Verify not found when no match
	foundIdentity, found, err = selected.Identity.Search().
		Direct("none").
		Indirect("none").
		Find()
	g.Expect(err).To(BeNil())
	g.Expect(found).To(BeFalse())
}

// TestApplicationTag tests the Application.Select().Tag subresource
func TestApplicationTag(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Tags",
		Description: "Application for testing tag subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// LIST: Verify initially empty
	list, err := selected.Tag.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(0))

	// ADD: Add seeded tag 1
	err = selected.Tag.Add(1)
	g.Expect(err).To(BeNil())

	// LIST: Verify tag was added
	list, err = selected.Tag.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	g.Expect(list[0].ID).To(Equal(uint(1)))

	// ENSURE: Ensure tag 1 (should be idempotent)
	err = selected.Tag.Ensure(1)
	g.Expect(err).To(BeNil())

	// LIST: Verify still only one tag
	list, err = selected.Tag.List()
	g.Expect(err).To(BeNil())
	eq, report := cmp.Eq(
		[]api.TagRef{
			{ID: 1},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)

	// ADD: Add seeded tag 2
	err = selected.Tag.Add(2)
	g.Expect(err).To(BeNil())

	// LIST: Verify both tags
	list, err = selected.Tag.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		[]api.TagRef{
			{ID: 1},
			{ID: 2},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove tag 404
	err = selected.Tag.Delete(1)
	g.Expect(err).To(BeNil())

	// LIST: Verify no tags remain
	list, err = selected.Tag.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		[]api.TagRef{
			{ID: 2},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)
}

// TestApplicationTag tests the Application.Select().Tag subresource
func TestApplicationTagWithSource(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Tags",
		Description: "Application for testing tag subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)
	source := selected.Tag.Source("T")

	// LIST: Verify initially empty
	list, err := source.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(0))

	// ADD: Add seeded tag 1
	err = source.Add(1)
	g.Expect(err).To(BeNil())

	// LIST: Verify tag was added
	list, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report := cmp.Eq(
		[]api.TagRef{
			{ID: 1, Source: "T"},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)

	// ENSURE: Ensure tag 1 (should be idempotent)
	err = source.Ensure(1)
	g.Expect(err).To(BeNil())

	// LIST: Verify still only one tag
	list, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		[]api.TagRef{
			{ID: 1, Source: "T"},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)

	// ADD: Add seeded tag 2
	err = source.Add(2)
	g.Expect(err).To(BeNil())

	// LIST: Verify both tags
	list, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		[]api.TagRef{
			{ID: 1, Source: "T"},
			{ID: 2, Source: "T"},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)

	// REPLACE: Replace T with only tags 4,5
	err = source.Replace([]uint{4, 5})
	g.Expect(err).To(BeNil())

	// LIST: Verify tag 1,2,3 and 404 remains
	list, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		[]api.TagRef{
			{ID: 4, Source: "T"},
			{ID: 5, Source: "T"},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove tag 404
	err = selected.Tag.Delete(5)
	g.Expect(err).To(BeNil())

	// LIST: Verify no tags remain
	list, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		[]api.TagRef{
			{ID: 4, Source: "T"},
		},
		list,
		".Name")
	g.Expect(eq).To(BeTrue(), report)
}

// TestApplicationAssessment tests the Application.Select().Assessment subresource
func TestApplicationAssessment(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Assessment",
		Description: "Application for testing assessment subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// CREATE: Create an assessment
	assessment := &api.Assessment{
		Questionnaire: api.Ref{
			ID:   1,
			Name: "Legacy Pathfinder"},
		Application: &api.Ref{
			ID:   app.ID,
			Name: app.Name,
		},
		Status: "started",
	}
	err = selected.Assessment.Create(assessment)
	g.Expect(err).To(BeNil())
	g.Expect(assessment.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment.ID)
	})

	// LIST: Verify assessment was created
	list, err := selected.Assessment.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(assessment, &list[0])
	g.Expect(eq).To(BeTrue(), report)
}

// TestApplicationAnalysis tests the Application.Select().Analysis subresource
func TestApplicationAnalysis(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Analysis",
		Description: "Application for testing analysis subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// CREATE: Create an analysis
	analysis := &api.Analysis{
		Effort:      100,
		Application: api.Ref{ID: app.ID},
	}
	err = selected.Analysis.Create(analysis)
	g.Expect(err).To(BeNil())
	g.Expect(analysis.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// LIST: List all analyses for the application
	analysisList, err := selected.Analysis.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(analysisList)).To(BeNumerically(">", 0))
	found := false
	for _, a := range analysisList {
		if a.ID == analysis.ID {
			found = true
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// GET: Retrieve the latest analysis
	retrieved, err := selected.Analysis.Get()
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(analysis.ID))

	// GET REPORT: Download the latest analysis report
	reportDest := "/tmp/test-app-analysis-report.tar.gz"
	defer os.Remove(reportDest)
	err = selected.Analysis.GetReport(reportDest)
	g.Expect(err).To(BeNil())

	// Verify the report file was created
	info, err := os.Stat(reportDest)
	g.Expect(err).To(BeNil())
	g.Expect(info.Size()).To(BeNumerically(">", 0))

	// LIST INSIGHTS: List insights for the latest analysis
	insights, err := selected.Analysis.ListInsights()
	g.Expect(err).To(BeNil())
	g.Expect(insights).NotTo(BeNil())

	// LIST DEPENDENCIES: List dependencies for the latest analysis
	dependencies, err := selected.Analysis.ListDependencies()
	g.Expect(err).To(BeNil())
	g.Expect(dependencies).NotTo(BeNil())
}

// TestApplicationManifest tests the Application.Select().Manifest subresource
func TestApplicationManifest(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Manifest",
		Description: "Application for testing manifest subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// CREATE: Create a manifest
	manifest := &api.Manifest{
		Application: api.Ref{ID: app.ID},
		Content:     api.Map{"test": "data"},
	}
	err = selected.Manifest.Create(manifest)
	g.Expect(err).To(BeNil())
	g.Expect(manifest.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Manifest.Delete(manifest.ID)
	})

	// GET: Retrieve the manifest
	retrieved, err := selected.Manifest.Get()
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(manifest.ID))
}

// TestApplicationManifestEncryption tests manifest encryption, decryption, and injection
func TestApplicationManifestEncryption(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Manifest Encryption",
		Description: "Application for testing manifest encryption",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// CREATE: Create a manifest with content and secrets
	manifest := &api.Manifest{
		Application: api.Ref{ID: app.ID},
		Content: api.Map{
			"name": "Test",
			"key":  "$(key)",
			"database": api.Map{
				"url":      "db.com",
				"user":     "$(user)",
				"password": "$(password)",
			},
			"description": "Connect using $(user) and $(password)",
		},
		Secret: api.Map{
			"key":      "ABCDEF",
			"user":     "Elmer",
			"password": "1234",
		},
	}
	originalSecret := api.Map{
		"key":      "ABCDEF",
		"user":     "Elmer",
		"password": "1234",
	}

	err = selected.Manifest.Create(manifest)
	g.Expect(err).To(BeNil())
	g.Expect(manifest.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Manifest.Delete(manifest.ID)
	})

	// Verify Content is unchanged
	eq, report := cmp.Eq(
		api.Map{
			"name": "Test",
			"key":  "$(key)",
			"database": api.Map{
				"url":      "db.com",
				"user":     "$(user)",
				"password": "$(password)",
			},
			"description": "Connect using $(user) and $(password)",
		},
		manifest.Content)
	g.Expect(eq).To(BeTrue(), report)

	// Verify Secret is encrypted (should NOT match original)
	eq, _ = cmp.Eq(originalSecret, manifest.Secret)
	g.Expect(eq).To(BeFalse(), "Secret should be encrypted after create")

	// GET: Retrieve with default (encrypted)
	encrypted, err := selected.Manifest.Get()
	g.Expect(err).To(BeNil())
	g.Expect(encrypted).NotTo(BeNil())

	// Verify Content is unchanged
	eq, report = cmp.Eq(manifest.Content, encrypted.Content)
	g.Expect(eq).To(BeTrue(), report)

	// Verify Secret is still encrypted
	eq, report = cmp.Eq(manifest.Secret, encrypted.Secret)
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve with Decrypted param
	decrypted, err := selected.Manifest.Get(binding.Param{Key: api.Decrypted, Value: "1"})
	g.Expect(err).To(BeNil())
	g.Expect(decrypted).NotTo(BeNil())

	// Verify Content is unchanged
	eq, report = cmp.Eq(manifest.Content, decrypted.Content)
	g.Expect(eq).To(BeTrue(), report)

	// Verify Secret is decrypted (matches original)
	eq, report = cmp.Eq(originalSecret, decrypted.Secret)
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve with Decrypted and Injected params
	injected, err := selected.Manifest.Get(
		binding.Param{Key: api.Decrypted, Value: "1"},
		binding.Param{Key: api.Injected, Value: "1"})
	g.Expect(err).To(BeNil())
	g.Expect(injected).NotTo(BeNil())

	// Verify Content has secrets injected
	expectedInjected := api.Map{
		"name": "Test",
		"key":  "ABCDEF",
		"database": api.Map{
			"url":      "db.com",
			"user":     "Elmer",
			"password": "1234",
		},
		"description": "Connect using Elmer and 1234",
	}
	eq, report = cmp.Eq(expectedInjected, injected.Content)
	g.Expect(eq).To(BeTrue(), report)

	// Verify Secret is decrypted in injected response (same as original)
	eq, report = cmp.Eq(originalSecret, injected.Secret)
	g.Expect(eq).To(BeTrue(), report)
}

// TestApplicationFact tests the Application.Select().Fact subresource
func TestApplicationFact(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Facts",
		Description: "Application for testing fact subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// LIST: Verify initially empty
	facts, err := selected.Fact.List()
	g.Expect(err).To(BeNil())
	eq, report := cmp.Eq(api.Map{}, facts)
	g.Expect(eq).To(BeTrue(), report)

	// SET: Set a fact
	err = selected.Fact.Set("test-key", "test-value")
	g.Expect(err).To(BeNil())

	// LIST: Verify fact was added
	facts, err = selected.Fact.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"test-key": "test-value",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)

	// SET: Add another fact
	err = selected.Fact.Set("test-key-2", "test-value-2")
	g.Expect(err).To(BeNil())

	// LIST: Verify both facts
	facts, err = selected.Fact.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"test-key":   "test-value",
			"test-key-2": "test-value-2",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Delete a fact
	err = selected.Fact.Delete("test-key")
	g.Expect(err).To(BeNil())

	// LIST: Verify deletion
	facts, err = selected.Fact.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"test-key-2": "test-value-2",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)
}

// TestApplicationFactWithSource tests the Application.Select().Fact subresource with source
func TestApplicationFactWithSource(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Facts",
		Description: "Application for testing fact subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)
	source := selected.Fact.Source("F")

	// LIST: Verify initially empty
	facts, err := source.List()
	g.Expect(err).To(BeNil())
	eq, report := cmp.Eq(api.Map{}, facts)
	g.Expect(eq).To(BeTrue(), report)

	// SET: Set a fact
	err = source.Set("test-key", "test-value")
	g.Expect(err).To(BeNil())

	// LIST: Verify fact was added
	facts, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"test-key": "test-value",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)

	// SET: Update the fact
	err = source.Set("test-key", "updated-value")
	g.Expect(err).To(BeNil())

	// LIST: Verify update
	facts, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"test-key": "updated-value",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)

	// SET: Add another fact
	err = source.Set("test-key-2", api.Map{
		"nested": "data",
		"count":  42,
	})
	g.Expect(err).To(BeNil())

	// LIST: Verify both facts
	facts, err = source.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(facts)).To(Equal(2))

	// REPLACE: Replace all facts
	newFacts := api.Map{
		"new-key-1": "new-value-1",
		"new-key-2": "new-value-2",
	}
	err = source.Replace(newFacts)
	g.Expect(err).To(BeNil())

	// LIST: Verify replacement
	facts, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"new-key-1": "new-value-1",
			"new-key-2": "new-value-2",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Delete a fact
	err = source.Delete("new-key-1")
	g.Expect(err).To(BeNil())

	// LIST: Verify deletion
	facts, err = source.List()
	g.Expect(err).To(BeNil())
	eq, report = cmp.Eq(
		api.Map{
			"new-key-2": "new-value-2",
		},
		facts)
	g.Expect(eq).To(BeTrue(), report)

	// GET: Verify deleted fact is not found
	var value string
	err = source.Get("new-key-1", &value)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestApplicationBucket tests the Application.Select().Bucket subresource
func TestApplicationBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for testing
	app := &api.Application{
		Name:        "Test App for Bucket",
		Description: "Application for testing bucket subresource",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	g.Expect(app.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Get the selected application API
	selected := client.Application.Select(app.ID)

	// PUT: Upload a file to the bucket
	tmpFile := "/tmp/test-bucket-source.txt"
	testContent := []byte("This is test content for the bucket")
	err = os.WriteFile(tmpFile, testContent, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile)

	err = selected.Bucket.Put(tmpFile, "test-file.txt")
	g.Expect(err).To(BeNil())

	// GET: Download the file
	tmpDest := "/tmp/test-bucket-dest.txt"
	defer os.Remove(tmpDest)
	err = selected.Bucket.Get("test-file.txt", tmpDest)
	g.Expect(err).To(BeNil())
	content, err := os.ReadFile(tmpDest)
	g.Expect(err).To(BeNil())
	g.Expect(content).To(Equal(testContent))

	// PUT: Upload another file
	tmpFile2 := "/tmp/test-bucket-nested.txt"
	nestedContent := []byte("nested content")
	err = os.WriteFile(tmpFile2, nestedContent, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile2)

	err = selected.Bucket.Put(tmpFile2, "test-dir/nested-file.txt")
	g.Expect(err).To(BeNil())

	// GET: Download nested file
	tmpDest2 := "/tmp/test-bucket-nested-dest.txt"
	defer os.Remove(tmpDest2)
	err = selected.Bucket.Get("test-dir/nested-file.txt", tmpDest2)
	g.Expect(err).To(BeNil())
	content, err = os.ReadFile(tmpDest2)
	g.Expect(err).To(BeNil())
	g.Expect(content).To(Equal(nestedContent))

	// DELETE: Delete a file
	err = selected.Bucket.Delete("test-file.txt")
	g.Expect(err).To(BeNil())

	// GET: Verify deletion
	err = selected.Bucket.Get("test-file.txt", tmpDest)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
