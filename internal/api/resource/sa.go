package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// ServiceAccount REST resource.
type ServiceAccount api.ServiceAccount

// With converts model to REST resource.
func (r *ServiceAccount) With(m *model.ServiceAccount) {
	baseWith(&r.Resource, &m.Model)
	m = mustRedact(m)
	r.Subject = m.Subject
	r.Name = m.Name
	r.Roles = []Ref{}
	for _, role := range m.Roles {
		r.Roles = append(r.Roles, Ref{ID: role.ID, Name: role.Name})
	}
	r.Tokens = []Ref{}
	for _, token := range m.Tokens {
		r.Tokens = append(r.Tokens, ref(token.ID, &token))
	}
}

// Model converts REST resource to model.
func (r *ServiceAccount) Model() (m *model.ServiceAccount) {
	m = &model.ServiceAccount{
		Subject: r.Subject,
		Name:    r.Name,
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
