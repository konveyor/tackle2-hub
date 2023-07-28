package review

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
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
			if got.BusinessCriticality != r.BusinessCriticality {
				t.Errorf("Different response error. Got %v, expected %v", got.BusinessCriticality, r.BusinessCriticality)
			}
			if got.EffortEstimate != r.EffortEstimate {
				t.Errorf("Different response error. Got %v, expected %v", got.EffortEstimate, r.EffortEstimate)
			}
			if got.ProposedAction != r.ProposedAction {
				t.Errorf("Different response error. Got %v, expected %v", got.ProposedAction, r.ProposedAction)
			}
			if got.WorkPriority != r.WorkPriority {
				t.Errorf("Different response error. Got %v, expected %v", got.WorkPriority, r.WorkPriority)
			}
			if got.Comments != r.Comments {
				t.Errorf("Different response error. Got %v, expected %v", got.Comments, r.Comments)
			}
			if got.Application.Name != r.Application.Name {
				t.Errorf("Different response error. Got %v, expected %v", got.Application.Name, r.Application.Name)
			}

			// Update Application Name.
			r.Application.Name = "Updated " + r.Application.Name
			assert.Should(t, Review.Update(&r))

			// Find Review and check its parameters with the got(On Updation).
			got, err = Review.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.BusinessCriticality != r.BusinessCriticality {
				t.Errorf("Different response error. Got %v, expected %v", got.BusinessCriticality, r.BusinessCriticality)
			}
			if got.EffortEstimate != r.EffortEstimate {
				t.Errorf("Different response error. Got %v, expected %v", got.EffortEstimate, r.EffortEstimate)
			}
			if got.ProposedAction != r.ProposedAction {
				t.Errorf("Different response error. Got %v, expected %v", got.ProposedAction, r.ProposedAction)
			}
			if got.WorkPriority != r.WorkPriority {
				t.Errorf("Different response error. Got %v, expected %v", got.WorkPriority, r.WorkPriority)
			}
			if got.Comments != r.Comments {
				t.Errorf("Different response error. Got %v, expected %v", got.Comments, r.Comments)
			}

			// Delete Related Applications.
			assert.Must(t, Application.Delete(app.ID))
		})

		t.Run("Delete Review and its dependencies", func(t *testing.T) {

			// Delete Review.
			assert.Must(t, Review.Delete(r.ID))

			// Check if the review is present even after deletion or not.
			_, err := Review.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
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
		assert.Must(t, Application.Delete(review.ID))
		assert.Must(t, Review.Delete(review.ID))
	}
}
