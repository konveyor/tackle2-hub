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
	r.Permissions = []Ref{}
	for _, perm := range m.Permissions {
		r.Permissions = append(r.Permissions, Ref{ID: perm.ID, Name: perm.Name})
	}
}

// Model converts REST resource to model.
func (r *Role) Model() (m *model.Role) {
	m = &model.Role{
		Name: r.Name,
	}
	m.ID = r.ID
	for _, ref := range r.Permissions {
		m.Permissions = append(
			m.Permissions,
			model.Permission{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}
