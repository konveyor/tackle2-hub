package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Review API.
type Review struct {
	client *Client
}

// Create a Review.
func (h *Review) Create(r *api.Review) (err error) {
	err = h.client.Post(api.ReviewsRoute, &r)
	return
}

// Get a Review by ID.
func (h *Review) Get(id uint) (r *api.Review, err error) {
	r = &api.Review{}
	path := Path(api.ReviewRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Reviews.
func (h *Review) List() (list []api.Review, err error) {
	list = []api.Review{}
	err = h.client.Get(api.ReviewsRoute, &list)
	return
}

// Update a Review.
func (h *Review) Update(r *api.Review) (err error) {
	path := Path(api.ReviewRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Review.
func (h *Review) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ReviewRoute).Inject(Params{api.ID: id}))
	return
}

// Copy a Review.
func (h *Review) Copy(reviewID uint, appID uint) (err error) {
	copyRequest := api.CopyRequest{
		SourceReview:       reviewID,
		TargetApplications: []uint{appID},
	}
	err = h.client.Post(api.CopyRoute, copyRequest)
	return
}
