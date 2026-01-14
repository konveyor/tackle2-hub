package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Manifest REST resource.
type Manifest api.Manifest

// With updates the resource with the model.
func (r *Manifest) With(m *model.Manifest) {
	baseWith(&r.Resource, &m.Model)
	r.Content = m.Content
	r.Secret = m.Secret
	r.Application = Ref{ID: m.ApplicationID, Name: ""}
}

// Model builds a model.
func (r *Manifest) Model() (m *model.Manifest) {
	m = &model.Manifest{}
	m.ID = r.ID
	m.Content = r.Content
	m.Secret = r.Secret
	m.ApplicationID = r.Application.ID
	return
}
