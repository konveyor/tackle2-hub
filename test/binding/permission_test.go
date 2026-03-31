package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestPermission(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the permission to create
	permission := &api.Permission{
		Name:  "testpermission",
		Scope: "testscope",
	}

	// Get seeded
	seeded, err := client.Permission.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the permission
	err = client.Permission.Create(permission)
	g.Expect(err).To(BeNil())
	g.Expect(permission.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Permission.Delete(permission.ID)
	})

	// GET: List permissions
	list, err := client.Permission.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))

	// GET: Retrieve the permission and verify it matches
	retrieved, err := client.Permission.Get(permission.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(permission, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the permission
	permission.Scope = "updatedscope"

	err = client.Permission.Update(permission)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Permission.Get(permission.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(permission, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the permission
	err = client.Permission.Delete(permission.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Permission.Get(permission.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
