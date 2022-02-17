package addon

import (
	"errors"
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
func (h *Setting) Get(key string) (v interface{}, err error) {
	r := &api.Setting{}
	path := Params{api.Key: key}.inject(api.SettingRoot)
	err = h.client.Get(path, r)
	v = r.Value
	return
}

//
// Bool setting value.
func (h *Setting) Bool(key string) (b bool, err error) {
	v, err := h.Get(key)
	if err != nil {
		return
	}
	b, cast := v.(bool)
	if !cast {
		err = errors.New(key + " not <boolean>")
	}
	return
}

//
// Str setting value.
func (h *Setting) Str(key string) (s string, err error) {
	v, err := h.Get(key)
	if err != nil {
		return
	}
	s, cast := v.(string)
	if !cast {
		err = errors.New(key + " not <string>")
	}
	return
}

//
// Int setting value.
func (h *Setting) Int(key string) (n int, err error) {
	v, err := h.Get(key)
	if err != nil {
		return
	}
	n, cast := v.(int)
	if !cast {
		err = errors.New(key + " not <int>")
	}
	return
}
