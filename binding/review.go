package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Review API.
type Review struct {
	client *Client
}

//
// Create a Review.
func (h *Review) Create(r *api.Review) (err error) {
	err = h.client.Post(api.ReviewsRoot, &r)
	return
}

//
// Get a Review by ID.
func (h *Review) Get(id uint) (r *api.Review, err error) {
	r = &api.Review{}
	path := Path(api.ReviewRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Stakeholders.
func (h *Review) List() (list []api.Review, err error) {
	list = []api.Review{}
	err = h.client.Get(api.ReviewsRoot, &list)
	return
}

//
// Update a Review.
func (h *Review) Update(r *api.Review) (err error) {
	path := Path(api.ReviewRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a Review.
func (h *Review) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ReviewRoot).Inject(Params{api.ID: id}))
	return
}
