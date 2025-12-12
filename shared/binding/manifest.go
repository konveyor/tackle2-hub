package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Manifest API.
type Manifest struct {
	client *Client
}

// Create a Manifest.
func (h *Manifest) Create(r *api.Manifest) (err error) {
	err = h.client.Post(api.ManifestsRoute, &r)
	return
}

// Get a Manifest by ID.
// Params:
// Param{Key: Decrypted, Value: "1"}
// Param{Key: Injected, Value: "1"}
func (h *Manifest) Get(id uint, param ...Param) (r *api.Manifest, err error) {
	r = &api.Manifest{}
	path := Path(api.ManifestRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r, param...)
	return
}

// List Manifests.
// Params:
// Param{Key: Decrypted, Value: "1"}
// Param{Key: Injected, Value: "1"}
func (h *Manifest) List(param ...Param) (list []api.Manifest, err error) {
	list = []api.Manifest{}
	err = h.client.Get(api.ManifestsRoute, &list, param...)
	return
}

// Find Manifests with filter.
func (h *Manifest) Find(filter Filter) (list []api.Manifest, err error) {
	list = []api.Manifest{}
	err = h.client.Get(api.ManifestsRoute, &list, filter.Param())
	return
}

// Update a Manifest.
func (h *Manifest) Update(r *api.Manifest) (err error) {
	path := Path(api.ManifestRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Manifest.
func (h *Manifest) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ManifestRoute).Inject(Params{api.ID: id}))
	return
}
