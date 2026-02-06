package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestJobFunction(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the job function to create
	jobFunction := &api.JobFunction{
		Name: "Engineer",
	}

	// Get seeded
	seeded, err := client.JobFunction.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the job function
	err = client.JobFunction.Create(jobFunction)
	g.Expect(err).To(BeNil())
	g.Expect(jobFunction.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.JobFunction.Delete(jobFunction.ID)
	})

	// GET: List job functions
	list, err := client.JobFunction.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(jobFunction, list[len(seeded)])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the job function and verify it matches
	retrieved, err := client.JobFunction.Get(jobFunction.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(jobFunction, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the job function
	jobFunction.Name = "Senior Engineer"

	err = client.JobFunction.Update(jobFunction)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.JobFunction.Get(jobFunction.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(jobFunction, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the job function
	err = client.JobFunction.Delete(jobFunction.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.JobFunction.Get(jobFunction.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
