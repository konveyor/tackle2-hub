package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestRole(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the role to create
	role := &api.Role{
		Name: "testrole",
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
	g.Expect(retrieved.Name).To(Equal(role.Name))

	// UPDATE: Modify the role
	role.Name = "updatedrole"

	err = client.Role.Update(role)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Role.Get(role.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.Name).To(Equal("updatedrole"))

	// DELETE: Remove the role
	err = client.Role.Delete(role.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Role.Get(role.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
