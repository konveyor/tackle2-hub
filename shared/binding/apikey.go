package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// APIKey API.
type APIKey struct {
	client RestClient
}

// Create an APIKey.
func (h APIKey) Create(r *api.APIKey) (err error) {
	err = h.client.Post(api.AuthAPIKeysRoute, r)
	return
}

// Get an APIKey by ID.
func (h APIKey) Get(id uint) (r *api.APIKey, err error) {
	r = &api.APIKey{}
	path := Path(api.AuthAPIKeyIDRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List APIKeys.
func (h APIKey) List() (list []api.APIKey, err error) {
	list = []api.APIKey{}
	err = h.client.Get(api.AuthAPIKeysRoute, &list)
	return
}

// Delete an APIKey.
func (h APIKey) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.AuthAPIKeyIDRoute).Inject(Params{api.ID: id}))
	return
}
