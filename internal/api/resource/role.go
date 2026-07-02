package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Role REST resource.
type Role api.Role

// With converts model to REST resource.
func (r *Role) With(m *model.Role) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Scopes = m.Scopes
}

// Model converts REST resource to model.
func (r *Role) Model() (m *model.Role) {
	m = &model.Role{
		Name:   r.Name,
		Scopes: r.Scopes,
	}
	m.ID = r.ID
	return
}
