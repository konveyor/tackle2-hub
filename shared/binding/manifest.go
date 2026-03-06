package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Manifest API.
type Manifest struct {
	client    RestClient
	decrypted bool
	injected  bool
}

// Decrypted enables decryption.
// Returned resources with fields decrypted.
func (h Manifest) Decrypted() (h2 Manifest) {
	h2 = Manifest{
		client:    h.client,
		injected:  h.injected,
		decrypted: true,
	}
	return
}

// Injected enables injection.
// Returned resources with secrets injected into the content.
func (h Manifest) Injected() (h2 Manifest) {
	h2 = Manifest{
		client:    h.client,
		decrypted: h.decrypted,
		injected:  true,
	}
	return
}

// Create a Manifest.
func (h Manifest) Create(r *api.Manifest) (err error) {
	err = h.client.Post(api.ManifestsRoute, r)
	return
}

// Get a Manifest by ID.
func (h Manifest) Get(id uint) (r *api.Manifest, err error) {
	r = &api.Manifest{}
	p := h.params()
	path := Path(api.ManifestRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r, p...)
	return
}

// List Manifests.
func (h Manifest) List() (list []api.Manifest, err error) {
	list = []api.Manifest{}
	p := h.params()
	err = h.client.Get(api.ManifestsRoute, &list, p...)
	return
}

// Find Manifests with filter.
func (h Manifest) Find(filter Filter) (list []api.Manifest, err error) {
	list = []api.Manifest{}
	p := h.params(filter)
	err = h.client.Get(api.ManifestsRoute, &list, p...)
	return
}

// Update a Manifest.
func (h Manifest) Update(r *api.Manifest) (err error) {
	path := Path(api.ManifestRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Manifest.
func (h Manifest) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ManifestRoute).Inject(Params{api.ID: id}))
	return
}

// params returns parameters.
func (h Manifest) params(filter ...Filter) (param []Param) {
	if h.decrypted {
		param = append(
			param, Param{
				Key:   api.Decrypted,
				Value: "1",
			})
	}
	if h.injected {
		param = append(
			param, Param{
				Key:   api.Injected,
				Value: "1",
			})
	}
	for _, f := range filter {
		param = append(param, f.Param())
	}
	return
}
