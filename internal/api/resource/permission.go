package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Permission REST resource.
type Permission api.Permission

// With converts model to REST resource.
func (r *Permission) With(m *model.Permission) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Noun = m.Resource
	r.Verb = m.Verb
	r.Scope = m.Scope
}

// Model converts REST resource to model.
func (r *Permission) Model() (m *model.Permission) {
	m = &model.Permission{
		Name:     r.Name,
		Scope:    r.Scope,
		Resource: r.Noun,
		Verb:     r.Verb,
	}
	m.ID = r.ID
	return
}
