package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Setting API.
type Setting struct {
	client *Client
}

// Get a setting by key.
func (h *Setting) Get(key string, v any) (err error) {
	path := Path(api2.SettingRoute).Inject(Params{api2.Key: key})
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
func (h *Setting) Create(r *api2.Setting) (err error) {
	err = h.client.Post(api2.SettingsRoute, &r)
	return
}

// List Settings.
func (h *Setting) List() (list []api2.Setting, err error) {
	list = []api2.Setting{}
	err = h.client.Get(api2.SettingsRoute, &list)
	return
}

// Update a Setting.
func (h *Setting) Update(r *api2.Setting) (err error) {
	path := Path(api2.SettingRoute).Inject(Params{api2.Key: r.Key})
	err = h.client.Put(path, r)
	return
}

// Delete a Setting.
func (h *Setting) Delete(key string) (err error) {
	path := Path(api2.SettingRoute).Inject(Params{api2.Key: key})
	err = h.client.Delete(path)
	return
}
