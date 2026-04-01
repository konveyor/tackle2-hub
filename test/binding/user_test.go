package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestUser(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "rosebud"

	// Create permissions for roles
	permission1 := &api.Permission{
		Name:  "read:users",
		Scope: "users:read",
	}
	err := client.Permission.Create(permission1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Permission.Delete(permission1.ID)
	})

	permission2 := &api.Permission{
		Name:  "write:users",
		Scope: "users:write",
	}
	err = client.Permission.Create(permission2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Permission.Delete(permission2.ID)
	})

	// Create roles for the user to reference
	role1 := &api.Role{
		Name: "admin",
		Permissions: []api.Ref{
			{ID: permission1.ID},
			{ID: permission2.ID},
		},
	}
	err = client.Role.Create(role1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Role.Delete(role1.ID)
	})

	role2 := &api.Role{
		Name: "viewer",
		Permissions: []api.Ref{
			{ID: permission1.ID},
		},
	}
	err = client.Role.Create(role2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Role.Delete(role2.ID)
	})

	// Define the user to create with password and roles
	user := &api.User{
		Name:     "testuser",
		Password: password,
		Email:    "testuser@example.com",
		Roles: []api.Ref{
			{ID: role1.ID},
			{ID: role2.ID},
		},
	}

	// Get seeded
	seeded, err := client.User.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the user
	err = client.User.Create(user)
	g.Expect(err).To(BeNil())
	g.Expect(user.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.User.Delete(user.ID)
	})

	// GET: List users
	list, err := client.User.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))

	// GET: Retrieve the user and verify it matches
	retrieved, err := client.User.Decrypted().Get(user.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())

	// Password should be decrypted when using Decrypted()
	g.Expect(retrieved.Password).To(Equal(password))

	// Verify basic fields
	g.Expect(retrieved.Name).To(Equal(user.Name))
	g.Expect(retrieved.UUID).ToNot(BeZero()) // assigned.
	g.Expect(retrieved.Email).To(Equal(user.Email))

	// Verify roles are associated
	g.Expect(len(retrieved.Roles)).To(Equal(2))
	g.Expect(retrieved.Roles).To(ContainElement(api.Ref{ID: role1.ID, Name: role1.Name}))
	g.Expect(retrieved.Roles).To(ContainElement(api.Ref{ID: role2.ID, Name: role2.Name}))

	// UPDATE: Modify the user
	user.Email = "newemail@example.com"
	user.Password = "newpassword456"

	err = client.User.Update(user)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.User.Decrypted().Get(user.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.Email).To(Equal(user.Email))
	g.Expect(updated.Password).To(Equal(user.Password))

	// DELETE: Remove the user
	err = client.User.Delete(user.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.User.Get(user.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
