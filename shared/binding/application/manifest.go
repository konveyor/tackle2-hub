package application

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Manifest sub-resource API.
type Manifest struct {
	client    client.RestClient
	decrypted bool
	injected  bool
	appId     uint
}

// Decrypted enables decryption.
// Returned resources with fields decrypted.
func (h Manifest) Decrypted() (h2 Manifest) {
	h2 = Manifest{
		client:    h.client,
		appId:     h.appId,
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
		appId:     h.appId,
		decrypted: h.decrypted,
		injected:  true,
	}
	return
}

// Create manifest.
func (h Manifest) Create(r *api.Manifest) (err error) {
	path := client.Path(api.ManifestsRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Post(path, r)
	return
}

// Get returns the LATEST manifest.
func (h Manifest) Get() (r *api.Manifest, err error) {
	r = &api.Manifest{}
	p := h.params()
	path := client.Path(api.AppManifestRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, r, p...)
	return
}

// params returns parameters.
func (h Manifest) params(filter ...client.Filter) (param []client.Param) {
	if h.decrypted {
		param = append(
			param, client.Param{
				Key:   api.Decrypted,
				Value: "1",
			})
	}
	if h.injected {
		param = append(
			param, client.Param{
				Key:   api.Injected,
				Value: "1",
			})
	}
	for _, f := range filter {
		param = append(param, f.Param())
	}
	return
}
