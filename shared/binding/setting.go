package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Setting API.
type Setting struct {
	client *Client
}

// Get a setting by key.
func (h *Setting) Get(key string, v any) (err error) {
	path := Path(api.SettingRoute).Inject(Params{api.Key: key})
	err = h.client.Get(path, v)
	return
}

// Bool setting value.
func (h *Setting) Bool(key string) (b bool, err error) {
	err = h.Get(key, &b)
	return
}

// Str setting value.
func (h *Setting) Str(key string) (s string, err error) {
	err = h.Get(key, &s)
	return
}

// Int setting value.
func (h *Setting) Int(key string) (n int, err error) {
	err = h.Get(key, &n)
	return
}

// Create a Setting.
func (h *Setting) Create(r *api.Setting) (err error) {
	err = h.client.Post(api.SettingsRoute, &r)
	return
}

// List Settings.
func (h *Setting) List() (list []api.Setting, err error) {
	list = []api.Setting{}
	err = h.client.Get(api.SettingsRoute, &list)
	return
}

// Update a Setting.
func (h *Setting) Update(r *api.Setting) (err error) {
	path := Path(api.SettingRoute).Inject(Params{api.Key: r.Key})
	err = h.client.Put(path, r)
	return
}

// Delete a Setting.
func (h *Setting) Delete(key string) (err error) {
	path := Path(api.SettingRoute).Inject(Params{api.Key: key})
	err = h.client.Delete(path)
	return
}
