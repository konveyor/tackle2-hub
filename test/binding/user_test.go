package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestUser(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the user to create
	user := &api.User{
		Name:     "testuser",
		Password: "testpassword",
		Email:    "testuser@example.com",
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
	g.Expect(retrieved.Name).To(Equal(user.Name))
	g.Expect(retrieved.Email).To(Equal(user.Email))

	// UPDATE: Modify the user
	user.Email = "newemail@example.com"

	err = client.User.Update(user)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.User.Get(user.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.Email).To(Equal("newemail@example.com"))

	// DELETE: Remove the user
	err = client.User.Delete(user.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.User.Get(user.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
