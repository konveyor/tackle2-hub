package resource

import (
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Token REST resource.
type Token api.Token

// With converts model to REST resource.
func (r *Token) With(m *model.Token) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.AuthId = m.AuthId
	r.Subject = m.Subject
	r.Scopes = m.Scopes
	r.Issued = m.Issued
	r.Lifespan = int(time.Now().Sub(r.Expiration) / time.Hour)
	r.Expiration = m.Expiration
	r.Grant = refPtr(m.GrantID, m.Grant)
	r.Task = refPtr(m.TaskID, m.Task)
	r.User = refPtr(m.UserID, m.User)
	r.IdpIdentity = refPtr(m.IdpIdentityID, m.IdpIdentity)
}
