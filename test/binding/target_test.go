package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTarget(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the target to create
	target := &api.Target{
		Name: "Test Target",
		Image: api.Ref{
			ID:   1,
			Name: "./data/image.svg",
		},
		Description: "Test target description",
	}

	// Get seeded.
	seeded, err := client.Target.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the target
	err = client.Target.Create(target)
	g.Expect(err).To(BeNil())
	g.Expect(target.ID).NotTo(BeZero())

	defer func() {
		_ = client.Target.Delete(target.ID)
	}()

	// GET: List targets
	list, err := client.Target.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(target, list[len(seeded)])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the target and verify it matches
	retrieved, err := client.Target.Get(target.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(target, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the target
	target.Name = "Updated Test Target"
	target.Description = "Updated test target description"

	err = client.Target.Update(target)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Target.Get(target.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(target, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the target
	err = client.Target.Delete(target.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Target.Get(target.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
