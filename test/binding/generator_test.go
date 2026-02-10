package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestGenerator(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create identity for the generator repository
	identity := &api.Identity{
		Name:     "generator-git-identity",
		Kind:     "git",
		User:     "git-user",
		Password: "git-password",
	}
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// Define the generator to create
	generator := &api.Generator{
		Kind:        "base",
		Name:        "Test Generator",
		Description: "This is a test generator",
		Repository: &api.Repository{
			Kind:   "git",
			URL:    "https://github.com/konveyor/tackle2-hub",
			Branch: "main",
		},
		Identity: &api.Ref{
			ID: identity.ID,
		},
		Params: api.Map{
			"p1": "v1",
			"p2": "v2",
		},
		Values: api.Map{
			"p1": "v1",
			"p2": "v2",
		},
	}

	// Get seeded.
	seeded, err := client.Generator.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the generator
	err = client.Generator.Create(generator)
	g.Expect(err).To(BeNil())
	g.Expect(generator.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Generator.Delete(generator.ID)
	})

	// GET: List generators
	list, err := client.Generator.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(generator, list[len(seeded)], "Identity.Name")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the generator and verify it matches
	retrieved, err := client.Generator.Get(generator.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(generator, retrieved, "Identity.Name")
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the generator
	generator.Name = "Updated Test Generator"
	generator.Description = "Updated generator description"
	generator.Params = api.Map{
		"p1": "updated-v1",
		"p3": "v3",
	}

	err = client.Generator.Update(generator)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Generator.Get(generator.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(generator, updated, "UpdateUser", "Identity.Name")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the generator
	err = client.Generator.Delete(generator.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Generator.Get(generator.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
