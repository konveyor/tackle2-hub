package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Setting REST resource.
type Setting api.Setting

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
