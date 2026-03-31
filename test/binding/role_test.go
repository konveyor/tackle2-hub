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

	// Create permissions for the role to reference
	permission1 := &api.Permission{
		Name:  "read:applications",
		Scope: "applications:read",
	}
	err := client.Permission.Create(permission1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Permission.Delete(permission1.ID)
	})

	permission2 := &api.Permission{
		Name:  "write:applications",
		Scope: "applications:write",
	}
	err = client.Permission.Create(permission2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Permission.Delete(permission2.ID)
	})

	// Define the role to create with permissions
	role := &api.Role{
		Name: "testrole",
		Permissions: []api.Ref{
			{ID: permission1.ID},
			{ID: permission2.ID},
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
	g.Expect(retrieved.Permissions).To(ContainElement(api.Ref{ID: permission1.ID, Name: permission1.Name}))
	g.Expect(retrieved.Permissions).To(ContainElement(api.Ref{ID: permission2.ID, Name: permission2.Name}))

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
