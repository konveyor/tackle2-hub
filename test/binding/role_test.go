package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestRole(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get available scopes from the hub
	scopes, err := client.Scope.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(scopes)).Should(BeNumerically(">=", 2))

	// Get seeded roles
	seeded, err := client.Role.List()
	g.Expect(err).To(BeNil())

	// Define the role to create with scopes
	role := &api.Role{
		Name: "testrole",
		Scopes: []string{
			scopes[0].Name,
			scopes[1].Name,
		},
	}

	// CREATE: Create the role
	err = client.Role.Create(role)
	g.Expect(err).To(BeNil())
	g.Expect(role.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Role.Delete(role.ID)
	})

	// GET: List roles
	list, err := client.Role.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))

	// GET: Retrieve the role and verify it matches
	retrieved, err := client.Role.Get(role.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(role, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// Verify scopes are associated
	g.Expect(len(retrieved.Scopes)).To(Equal(2))
	g.Expect(retrieved.Scopes).To(ContainElement(scopes[0].Name))
	g.Expect(retrieved.Scopes).To(ContainElement(scopes[1].Name))

	// UPDATE: Modify the role
	role.Name = "updatedrole"
	role.Scopes = []string{scopes[0].Name}

	err = client.Role.Update(role)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Role.Get(role.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(role, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the role
	err = client.Role.Delete(role.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Role.Get(role.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
