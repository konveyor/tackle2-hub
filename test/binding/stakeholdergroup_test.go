package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestStakeholderGroup(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the stakeholder group to create
	group := &api.StakeholderGroup{
		Name:        "Engineering",
		Description: "Engineering team.",
	}

	// CREATE: Create the stakeholder group
	err := client.StakeholderGroup.Create(group)
	g.Expect(err).To(BeNil())
	g.Expect(group.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.StakeholderGroup.Delete(group.ID)
	})

	// GET: List stakeholder groups
	list, err := client.StakeholderGroup.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(group, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the stakeholder group and verify it matches
	retrieved, err := client.StakeholderGroup.Get(group.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(group, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the stakeholder group
	group.Name = "Engineering Updated"
	group.Description = "Updated engineering team description."

	err = client.StakeholderGroup.Update(group)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.StakeholderGroup.Get(group.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(group, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the stakeholder group
	err = client.StakeholderGroup.Delete(group.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.StakeholderGroup.Get(group.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
