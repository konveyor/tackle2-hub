package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Review API.
type Review struct {
	client *Client
}

// Create a Review.
func (h *Review) Create(r *api2.Review) (err error) {
	err = h.client.Post(api2.ReviewsRoute, &r)
	return
}

// Get a Review by ID.
func (h *Review) Get(id uint) (r *api2.Review, err error) {
	r = &api2.Review{}
	path := Path(api2.ReviewRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Reviews.
func (h *Review) List() (list []api2.Review, err error) {
	list = []api2.Review{}
	err = h.client.Get(api2.ReviewsRoute, &list)
	return
}

// Update a Review.
func (h *Review) Update(r *api2.Review) (err error) {
	path := Path(api2.ReviewRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Review.
func (h *Review) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.ReviewRoute).Inject(Params{api2.ID: id}))
	return
}

// Copy a Review.
func (h *Review) Copy(reviewID uint, appID uint) (err error) {
	copyRequest := api2.CopyRequest{
		SourceReview:       reviewID,
		TargetApplications: []uint{appID},
	}
	err = h.client.Post(api2.CopyRoute, copyRequest)
	return
}
