package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Manifest API.
type Manifest struct {
	client *Client
}

// Create a Manifest.
func (h *Manifest) Create(r *api2.Manifest) (err error) {
	err = h.client.Post(api2.ManifestsRoute, &r)
	return
}

// Get a Manifest by ID.
// Params:
// Param{Key: Decrypted, Value: "1"}
// Param{Key: Injected, Value: "1"}
func (h *Manifest) Get(id uint, param ...Param) (r *api2.Manifest, err error) {
	r = &api2.Manifest{}
	path := Path(api2.ManifestRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r, param...)
	return
}

// List Manifests.
// Params:
// Param{Key: Decrypted, Value: "1"}
// Param{Key: Injected, Value: "1"}
func (h *Manifest) List(param ...Param) (list []api2.Manifest, err error) {
	list = []api2.Manifest{}
	err = h.client.Get(api2.ManifestsRoute, &list, param...)
	return
}

// Find Manifests with filter.
func (h *Manifest) Find(filter Filter) (list []api2.Manifest, err error) {
	list = []api2.Manifest{}
	err = h.client.Get(api2.ManifestsRoute, &list, filter.Param())
	return
}

// Update a Manifest.
func (h *Manifest) Update(r *api2.Manifest) (err error) {
	path := Path(api2.ManifestRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Manifest.
func (h *Manifest) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.ManifestRoute).Inject(Params{api2.ID: id}))
	return
}
