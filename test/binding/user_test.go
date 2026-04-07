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
		Userid:   "testuser",
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
	retrieved, err := client.User.Get(user.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())

	// Password should be hashed (bcrypt format)
	g.Expect(retrieved.Password).NotTo(BeEmpty())
	g.Expect(retrieved.Password).To(HavePrefix("bcrypt:"))
	g.Expect(retrieved.Password).NotTo(Equal(password)) // Should not equal plaintext

	// Verify basic fields
	g.Expect(retrieved.Userid).To(Equal(user.Userid))
	g.Expect(retrieved.UUID).ToNot(BeZero()) // assigned.
	g.Expect(retrieved.Email).To(Equal(user.Email))

	// Verify roles are associated
	g.Expect(len(retrieved.Roles)).To(Equal(2))
	g.Expect(retrieved.Roles).To(ContainElement(api.Ref{ID: role1.ID, Name: role1.Name}))
	g.Expect(retrieved.Roles).To(ContainElement(api.Ref{ID: role2.ID, Name: role2.Name}))

	// UPDATE: Modify the user
	user.Email = "newemail@example.com"
	newPassword := "newpassword456"
	user.Password = newPassword

	err = client.User.Update(user)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.User.Get(user.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.Email).To(Equal(user.Email))
	// Password should be hashed, not plaintext
	g.Expect(updated.Password).NotTo(BeEmpty())
	g.Expect(updated.Password).To(HavePrefix("bcrypt:"))
	g.Expect(updated.Password).NotTo(Equal(newPassword)) // Should not equal plaintext

	// DELETE: Remove the user
	err = client.User.Delete(user.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.User.Get(user.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

func TestUserPasswordTruncation(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a password exactly 72 bytes
	password72 := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890123456789"
	g.Expect(len(password72)).To(Equal(72))

	// Create a password longer than 72 bytes
	password80 := password72 + "12345678"
	g.Expect(len(password80)).To(Equal(80))

	// CREATE: User with 72-byte password (should succeed)
	user72 := &api.User{
		Userid:   "user72",
		Password: password72,
		Email:    "user72@example.com",
	}
	err := client.User.Create(user72)
	g.Expect(err).To(BeNil())
	g.Expect(user72.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.User.Delete(user72.ID)
	})

	// CREATE: User with 80-byte password (should fail validation)
	user80 := &api.User{
		Userid:   "user80",
		Password: password80,
		Email:    "user80@example.com",
	}
	err = client.User.Create(user80)
	g.Expect(err).NotTo(BeNil()) // Should fail due to max=72 validation

	// GET: Verify 72-byte password is hashed
	retrieved72, err := client.User.Get(user72.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved72.Password).NotTo(BeEmpty())
	g.Expect(retrieved72.Password).To(HavePrefix("bcrypt:"))
	g.Expect(retrieved72.Password).NotTo(Equal(password72)) // Should be hashed, not plaintext

	// UPDATE: Update user with password > 72 bytes (should fail validation)
	user72.Password = password80
	err = client.User.Update(user72)
	g.Expect(err).NotTo(BeNil()) // Should fail due to max=72 validation

	// UPDATE: Update user with valid 72-byte password (should succeed)
	newPassword72 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz0123456789"
	g.Expect(len(newPassword72)).To(Equal(72))
	user72.Password = newPassword72
	err = client.User.Update(user72)
	g.Expect(err).To(BeNil())

	// GET: Verify updated password is hashed
	updated, err := client.User.Get(user72.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated.Password).NotTo(BeEmpty())
	g.Expect(updated.Password).To(HavePrefix("bcrypt:"))
	g.Expect(updated.Password).NotTo(Equal(newPassword72)) // Should be hashed, not plaintext
}
