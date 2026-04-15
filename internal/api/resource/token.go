package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Token REST resource.
type Token api.Token

// With converts model to REST resource.
func (r *Token) With(m *model.Token) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.Subject = m.Subject
	r.Scopes = m.Scopes
	r.Issued = m.Issued
	r.Expiration = m.Expiration
	r.User = refPtr(m.UserID, m.User)
}
