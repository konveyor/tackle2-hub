package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// ConfigMap API.
type ConfigMap struct {
	client RestClient
}

// Get a ConfigMap by name.
func (h ConfigMap) Get(name string) (r *api.ConfigMap, err error) {
	r = &api.ConfigMap{}
	path := Path(api.ConfigMapRoute).Inject(Params{api.Name: name})
	err = h.client.Get(path, r)
	return
}

// GetKey gets a specific key value from a ConfigMap by name and key.
func (h ConfigMap) GetKey(name, key string) (value string, err error) {
	path := Path(api.ConfigMapKeyRoute).Inject(Params{
		api.Name: name,
		api.Key:  key,
	})
	err = h.client.Get(path, &value)
	return
}

// List ConfigMaps.
func (h ConfigMap) List() (list []api.ConfigMap, err error) {
	list = []api.ConfigMap{}
	err = h.client.Get(api.ConfigMapsRoute, &list)
	return
}
