package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestStakeholder(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create stakeholder group for the stakeholder to reference
	stakeholderGroup := &api.StakeholderGroup{
		Name:        "Engineering Group",
		Description: "Engineering team group",
	}
	err := client.StakeholderGroup.Create(stakeholderGroup)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.StakeholderGroup.Delete(stakeholderGroup.ID)
	})

	// Create business service for the stakeholder to reference
	businessService := &api.BusinessService{
		Name:        "HR Services",
		Description: "Human resources services",
	}
	err = client.BusinessService.Create(businessService)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.BusinessService.Delete(businessService.ID)
	})

	// Define the stakeholder to create
	stakeholder := &api.Stakeholder{
		Name:  "Alice",
		Email: "alice@acme.local",
		Groups: []api.Ref{
			{ID: stakeholderGroup.ID},
		},
		BusinessServices: []api.Ref{
			{ID: businessService.ID},
		},
		JobFunction: &api.Ref{
			ID: 1, // Use seeded job function
		},
	}

	// CREATE: Create the stakeholder
	err = client.Stakeholder.Create(stakeholder)
	g.Expect(err).To(BeNil())
	g.Expect(stakeholder.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Stakeholder.Delete(stakeholder.ID)
	})

	// GET: List stakeholders
	list, err := client.Stakeholder.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(stakeholder, list[0], "Groups.Name", "BusinessServices.Name", "JobFunction.Name")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the stakeholder and verify it matches
	retrieved, err := client.Stakeholder.Get(stakeholder.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(stakeholder, retrieved, "Groups.Name", "BusinessServices.Name", "JobFunction.Name")
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the stakeholder
	stakeholder.Name = "Alice Updated"
	stakeholder.Email = "alice.updated@acme.local"

	err = client.Stakeholder.Update(stakeholder)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Stakeholder.Get(stakeholder.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(stakeholder, updated, "UpdateUser", "Groups.Name", "BusinessServices.Name", "JobFunction.Name")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the stakeholder
	err = client.Stakeholder.Delete(stakeholder.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Stakeholder.Get(stakeholder.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
