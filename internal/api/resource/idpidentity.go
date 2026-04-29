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
	r.Issuer = m.Issuer
	r.Subject = m.Subject
	r.Userid = m.Userid
	r.Email = m.Email
	r.Expiration = m.Expiration
	r.LastAuthenticated = m.LastAuthenticated
	r.LastRefreshed = m.LastRefreshed
	r.Scopes = m.Scopes
	r.Roles = m.Roles
}

// Model converts REST resource to model.
func (r *IdpIdentity) Model() (m *model.IdpIdentity) {
	m = &model.IdpIdentity{
		Issuer:            r.Issuer,
		Subject:           r.Subject,
		Userid:            r.Userid,
		Email:             r.Email,
		Expiration:        r.Expiration,
		LastAuthenticated: r.LastAuthenticated,
		LastRefreshed:     r.LastRefreshed,
		Scopes:            r.Scopes,
		Roles:             r.Roles,
	}
	m.ID = r.ID
	return
}
