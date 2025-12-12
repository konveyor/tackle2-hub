package review

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestReviewCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("Review CRUD", func(t *testing.T) {

			// Create Application and Review.
			app := api.Application{
				Name:        r.Application.Name,
				Description: "Application for Review",
			}
			assert.Must(t, Application.Create(&app))

			// Update Application ID with the Sample Id.
			r.Application.ID = app.ID

			assert.Must(t, Review.Create(&r))

			// Get.
			got, err := Review.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare got values with expected values.
			AssertEqualReviews(t, got, r)

			// Update Comments and Effort Estimate.
			r.Comments = "Updated Comment " + r.Comments
			assert.Should(t, Review.Update(&r))

			// Find Review and check its parameters with the got(On Updation).
			got, err = Review.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check if the unchanged values remain same or not.
			AssertEqualReviews(t, got, r)

			// Delete Review.
			assert.Must(t, Review.Delete(r.ID))

			// Delete Related Applications.
			assert.Must(t, Application.Delete(app.ID))

			// Check if the review is present even after deletion or not.
			_, err = Review.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})

		t.Run("Copy Review", func(t *testing.T) {

			// Create Application and Review.
			srcApp := api.Application{
				Name:        r.Application.Name,
				Description: "Application for Review",
			}
			assert.Must(t, Application.Create(&srcApp))

			// Update Application ID with the Sample Id.
			r.Application.ID = srcApp.ID

			assert.Must(t, Review.Create(&r))

			// Create another application to copy
			destApp := api.Application{
				Name:        "New Application",
				Description: "Application for Review",
			}
			assert.Must(t, Application.Create(&destApp))

			err := Review.Copy(r.ID, destApp.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			destAppRef, _ := Application.Get(destApp.ID)
			gotReview, err := Review.Get(destAppRef.Review.ID)
			if err != nil {
				fmt.Println(err.Error())
				t.Errorf(err.Error())
			}

			// Check if the expcted review and got review is same.
			AssertEqualReviews(t, gotReview, r)

			// Delete Review.
			assert.Must(t, Review.Delete(r.ID))
			assert.Must(t, Review.Delete(gotReview.ID))

			// Delete Applications.
			assert.Must(t, Application.Delete(srcApp.ID))
			assert.Must(t, Application.Delete(destApp.ID))
		})
	}
}

func TestReviewList(t *testing.T) {
	createdReviews := []api.Review{}

	for _, r := range Samples {
		app := api.Application{
			Name:        r.Application.Name,
			Description: "Application for Review",
		}
		assert.Must(t, Application.Create(&app))

		r.Application.ID = app.ID
		assert.Should(t, Review.Create(&r))
		createdReviews = append(createdReviews, r)
	}

	// List Reviews.
	got, err := Review.List()
	if err != nil {
		t.Errorf(err.Error())
	}

	// check if created Reviews are in the list we got from Review.List()
	for _, createdReview := range createdReviews {
		found := false
		for _, retrievedReview := range got {
			if assert.FlatEqual(createdReview.ID, retrievedReview.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected review not found in the list: %v", createdReview)
		}
	}

	// Delete related reviews and applications.
	for _, review := range createdReviews {
		assert.Must(t, Application.Delete(review.Application.ID))
		assert.Must(t, Review.Delete(review.ID))
	}
}

func AssertEqualReviews(t *testing.T, got *api.Review, expected api.Review) {
	if got.BusinessCriticality != expected.BusinessCriticality {
		t.Errorf("Different Business Criticality Got %v, expected %v", got.BusinessCriticality, expected.BusinessCriticality)
	}
	if got.EffortEstimate != expected.EffortEstimate {
		t.Errorf("Different Effort Estimate Got %v, expected %v", got.EffortEstimate, expected.EffortEstimate)
	}
	if got.ProposedAction != expected.ProposedAction {
		t.Errorf("Different Proposed Action Got %v, expected %v", got.ProposedAction, expected.ProposedAction)
	}
	if got.WorkPriority != expected.WorkPriority {
		t.Errorf("Different Work Priority Got %v, expected %v", got.WorkPriority, expected.WorkPriority)
	}
}
