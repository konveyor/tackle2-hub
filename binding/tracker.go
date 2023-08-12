package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Tracker API.
type Tracker struct {
	client *Client
}

//
// Create a Tracker.
func (h *Tracker) Create(r *api.Tracker) (err error) {
	err = h.client.Post(api.TrackersRoot, &r)
	return
}

//
// Get a Tracker by ID.
func (h *Tracker) Get(id uint) (r *api.Tracker, err error) {
	r = &api.Tracker{}
	path := Path(api.TrackerRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Trackers.
func (h *Tracker) List() (list []api.Tracker, err error) {
	list = []api.Tracker{}
	err = h.client.Get(api.TrackersRoot, &list)
	return
}

//
// Update a Tracker.
func (h *Tracker) Update(r *api.Tracker) (err error) {
	path := Path(api.TrackerRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a Tracker.
func (h *Tracker) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TrackerRoot).Inject(Params{api.ID: id}))
	return
}
