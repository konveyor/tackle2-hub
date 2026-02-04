package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestBusinessService(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the business service to create
	businessService := &api.BusinessService{
		Name:        "Marketing",
		Description: "Marketing dept service.",
	}

	// CREATE: Create the business service
	err := client.BusinessService.Create(businessService)
	g.Expect(err).To(BeNil())
	g.Expect(businessService.ID).NotTo(BeZero())

	defer func() {
		_ = client.BusinessService.Delete(businessService.ID)
	}()

	// GET: List business services
	list, err := client.BusinessService.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(businessService, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the business service and verify it matches
	retrieved, err := client.BusinessService.Get(businessService.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(businessService, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the business service
	businessService.Name = "Marketing Updated"
	businessService.Description = "Updated marketing dept service description."

	err = client.BusinessService.Update(businessService)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.BusinessService.Get(businessService.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(businessService, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the business service
	err = client.BusinessService.Delete(businessService.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.BusinessService.Get(businessService.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
