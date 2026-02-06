package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestIdentity(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the identity to create
	identity := &api.Identity{
		Name:     "test-git-identity",
		Kind:     "git",
		User:     "test-user",
		Password: "test-password-123",
	}

	// CREATE: Create the identity
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	g.Expect(identity.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// GET: List identities
	list, err := client.Identity.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	list[0].Password = identity.Password // encrypted.
	eq, report := cmp.Eq(identity, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the identity and verify it matches
	retrieved, err := client.Identity.Get(identity.ID)
	retrieved.Password = identity.Password // encrypted.
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(identity, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the identity
	identity.Name = "updated-git-identity"
	identity.User = "updated-user"
	identity.Password = "updated-password-456"
	identity.Description = "Updated description"

	err = client.Identity.Update(identity)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Identity.Get(identity.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(identity, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the identity
	err = client.Identity.Delete(identity.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Identity.Get(identity.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestIdentityFind tests finding identities using filter
func TestIdentityFind(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create first identity
	direct := &api.Identity{
		Name: "direct",
		Kind: "Test",
	}
	err := client.Identity.Create(direct)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(direct.ID)
	})

	// Create second identity with different kind
	direct2 := &api.Identity{
		Name: "direct2",
		Kind: "Other",
	}
	err = client.Identity.Create(direct2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(direct2.ID)
	})

	// Create application with first identity
	application := &api.Application{
		Name:       "Test App for Identity Find",
		Identities: []api.IdentityRef{{ID: direct.ID}},
	}
	err = client.Application.Create(application)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

	// FIND: Find identity using filter for application.id and kind
	filter := binding.Filter{}
	filter.And("application.id").Eq(int(application.ID))
	filter.And("kind").Eq(direct.Kind)
	found, err := client.Identity.Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(found)).To(BeNumerically(">", 0), "Should find at least one identity")

	// Verify found identity is the correct one
	identity := found[0]
	g.Expect(identity.ID).To(Equal(direct.ID))
	g.Expect(identity.Kind).To(Equal(direct.Kind))
}
