package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// File REST resource.
type File api.File

// With updates the resource with the model.
func (r *File) With(m *model.File) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Path = m.Path
	r.Encoding = m.Encoding
	r.Expiration = m.Expiration
}

// Model builds a model.
func (r *File) Model() (m *model.File) {
	m = &model.File{}
	m.ID = r.ID
	m.Name = r.Name
	m.Path = r.Path
	m.Encoding = r.Encoding
	m.Expiration = r.Expiration
	return
}
