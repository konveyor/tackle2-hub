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

	permissions, err := client.Permission.List()
	g.Expect(len(permissions)).Should(BeNumerically(">=", 2))
	g.Expect(err).To(BeNil())

	// Define the role to create with permissions
	role := &api.Role{
		Name: "testrole",
		Permissions: []api.Ref{
			{ID: permissions[0].ID},
			{ID: permissions[1].ID},
		},
	}

	// Get seeded
	seeded, err := client.Role.List()
	g.Expect(err).To(BeNil())

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
	eq, report := cmp.Eq(role, retrieved, "Permissions.Name")
	g.Expect(eq).To(BeTrue(), report)

	// Verify permissions are associated
	g.Expect(len(retrieved.Permissions)).To(Equal(2))
	g.Expect(retrieved.Permissions).To(ContainElement(api.Ref{ID: permissions[0].ID, Name: permissions[0].Name}))
	g.Expect(retrieved.Permissions).To(ContainElement(api.Ref{ID: permissions[1].ID, Name: permissions[1].Name}))

	// UPDATE: Modify the role
	role.Name = "updatedrole"

	err = client.Role.Update(role)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Role.Get(role.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(role, updated, "UpdateUser", "Permissions.Name")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the role
	err = client.Role.Delete(role.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Role.Get(role.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
