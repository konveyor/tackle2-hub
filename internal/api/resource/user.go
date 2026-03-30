package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// User REST resource.
type User api.User

// With converts model to REST resource.
func (r *User) With(m *model.User) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Password = m.Password
	r.Email = m.Email
	r.Roles = []Ref{}
	for _, role := range m.Roles {
		r.Roles = append(r.Roles, Ref{ID: role.ID, Name: role.Name})
	}
}

// Model converts REST resource to model.
func (r *User) Model() (m *model.User) {
	m = &model.User{
		Name:     r.Name,
		Password: r.Password,
		Email:    r.Email,
	}
	m.ID = r.ID
	for _, ref := range r.Roles {
		m.Roles = append(
			m.Roles,
			model.Role{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}
