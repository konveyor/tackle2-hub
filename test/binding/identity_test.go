package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
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

	defer func() {
		_ = client.Identity.Delete(identity.ID)
	}()

	// GET: Retrieve the identity and verify it matches
	retrieved, err := client.Identity.Get(identity.ID)
	retrieved.Password = identity.Password // encrypted.
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(identity, retrieved)
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
