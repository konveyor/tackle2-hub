package application

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Manifest sub-resource API.
type Manifest struct {
	client *client.Client
	appId  uint
}

// Create manifest.
func (h Manifest) Create(r *api.Manifest) (err error) {
	path := client.Path(api.ManifestsRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Post(path, r)
	return
}

// Get returns the LATEST manifest.
// Params:
// Param{Key: Decrypted, Value: "1"}
// Param{Key: Injected, Value: "1"}
func (h Manifest) Get(param ...client.Param) (r *api.Manifest, err error) {
	r = &api.Manifest{}
	path := client.Path(api.AppManifestRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, r, param...)
	return
}
