package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// IdpIdentity REST resource.
type IdpIdentity api.IdpIdentity

// With converts model to REST resource.
func (r *IdpIdentity) With(m *model.IdpIdentity) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.Issuer = m.Issuer
	r.Subject = m.Subject
	r.Login = m.Login
	r.Name = m.Name
	r.Email = m.Email
	r.Expiration = m.Expiration
	r.LastAuthenticated = m.LastAuthenticated
	r.LastRefreshed = m.LastRefreshed
	r.Scopes = m.Scopes
	r.Tokens = []Ref{}
	for _, token := range m.Tokens {
		r.Tokens = append(r.Tokens, ref(token.ID, &token))
	}
}

// Model converts REST resource to model.
func (r *IdpIdentity) Model() (m *model.IdpIdentity) {
	m = &model.IdpIdentity{
		Kind:              r.Kind,
		Issuer:            r.Issuer,
		Subject:           r.Subject,
		Login:             r.Login,
		Name:              r.Name,
		Email:             r.Email,
		Expiration:        r.Expiration,
		LastAuthenticated: r.LastAuthenticated,
		LastRefreshed:     r.LastRefreshed,
		Scopes:            r.Scopes,
	}
	m.ID = r.ID
	return
}
