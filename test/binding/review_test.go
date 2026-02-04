package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestReview(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for the review to reference
	application := &api.Application{
		Name:        "Test Review App",
		Description: "Application for review testing",
	}
	err := client.Application.Create(application)
	g.Expect(err).To(BeNil())
	defer func() {
		_ = client.Application.Delete(application.ID)
	}()

	// Define the review to create
	review := &api.Review{
		BusinessCriticality: 1,
		EffortEstimate:      "small",
		ProposedAction:      "proceed",
		WorkPriority:        1,
		Comments:            "Initial review comments",
		Application: &api.Ref{
			ID:   application.ID,
			Name: application.Name,
		},
	}

	// CREATE: Create the review
	err = client.Review.Create(review)
	g.Expect(err).To(BeNil())
	g.Expect(review.ID).NotTo(BeZero())

	defer func() {
		_ = client.Review.Delete(review.ID)
	}()

	// GET: Retrieve the review and verify it matches
	retrieved, err := client.Review.Get(review.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(review, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the review
	review.BusinessCriticality = 2
	review.EffortEstimate = "medium"
	review.ProposedAction = "review"
	review.WorkPriority = 2
	review.Comments = "Updated review comments"

	err = client.Review.Update(review)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Review.Get(review.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(review, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the review
	err = client.Review.Delete(review.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Review.Get(review.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
