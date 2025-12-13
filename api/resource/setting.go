package resource

import "github.com/konveyor/tackle2-hub/model"

// Setting REST resource.
type Setting struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// With updates the resource with the model.
func (r *Setting) With(m *model.Setting) {
	r.Key = m.Key
	r.Value = m.Value
}

// Model builds a model.
func (r *Setting) Model() (m *model.Setting) {
	m = &model.Setting{Key: r.Key}
	m.Value = r.Value
	return
}
