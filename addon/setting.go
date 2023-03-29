package addon

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Setting API.
type Setting struct {
	// hub API client.
	client *Client
}

//
// Get a setting by key.
func (h *Setting) Get(key string, v interface{}) (err error) {
	path := Path(api.SettingRoot).Inject(Params{api.Key: key})
	err = h.client.Get(path, v)
	return
}

//
// Bool setting value.
func (h *Setting) Bool(key string) (b bool, err error) {
	err = h.Get(key, &b)
	return
}

//
// Str setting value.
func (h *Setting) Str(key string) (s string, err error) {
	err = h.Get(key, &s)
	return
}

//
// Int setting value.
func (h *Setting) Int(key string) (n int, err error) {
	err = h.Get(key, &n)
	return
}
