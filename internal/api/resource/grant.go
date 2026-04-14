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
	r.GrantId = m.GrantId
	r.ClientId = m.ClientId
	r.Subject = m.Subject
	r.Type = m.Type
	r.Scopes = m.Scopes
	r.Expiration = m.Expiration
}
