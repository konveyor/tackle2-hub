package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Grant REST resource.
type Grant api.Grant

// With converts model to REST resource.
func (r *Grant) With(m *model.Grant) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.AuthId = m.AuthId
	r.Subject = m.Subject
	r.Scopes = m.Scopes
	r.Issued = m.Issued
	r.Expiration = m.Expiration
	r.User = refPtr(m.UserID, m.User)
	r.IdpIdentity = refPtr(m.IdpIdentityID, m.IdpIdentity)
	r.IdpClient = refPtr(m.IdpClientID, m.IdpClient)
	r.Tokens = []Ref{}
	for i := range m.Tokens {
		r.Tokens = append(r.Tokens, ref(m.Tokens[i].ID, &m.Tokens[i]))
	}
}
