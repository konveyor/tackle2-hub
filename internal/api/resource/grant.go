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
	r.Subject = m.Subject
	r.Scopes = m.Scopes
	r.Issued = m.Issued
	r.Expiration = m.Expiration
}
