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
	r.Provider = m.Provider
	r.Subject = m.Subject
	r.RefreshToken = m.RefreshToken
	r.Expiration = m.Expiration
	r.LastAuthenticated = m.LastAuthenticated
	r.LastRefreshed = m.LastRefreshed
	r.User = &Ref{ID: m.UserID, Name: m.User.UserId}
}

// Model converts REST resource to model.
func (r *IdpIdentity) Model() (m *model.IdpIdentity) {
	m = &model.IdpIdentity{
		Provider:          r.Provider,
		Subject:           r.Subject,
		RefreshToken:      r.RefreshToken,
		Expiration:        r.Expiration,
		LastAuthenticated: r.LastAuthenticated,
		LastRefreshed:     r.LastRefreshed,
	}
	m.ID = r.ID
	if r.User != nil {
		m.UserID = r.User.ID
	}
	return
}
