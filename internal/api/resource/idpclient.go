package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// IdpClient REST resource.
type IdpClient api.IdpClient

// With converts model to REST resource.
func (r *IdpClient) With(m *model.IdpClient) {
	baseWith(&r.Resource, &m.Model)
	m = mustRedact(m)
	r.ClientId = m.ClientId
	r.Secret = m.Secret
	r.ApplicationType = m.ApplicationType
	r.Grants = m.Grants
	r.RedirectURIs = m.RedirectURIs
	r.Scopes = m.Scopes
	r.Tokens = []Ref{}
	for _, token := range m.Tokens {
		r.Tokens = append(r.Tokens, ref(token.ID, &token))
	}
}

// Model converts REST resource to model.
func (r *IdpClient) Model() (m *model.IdpClient) {
	m = &model.IdpClient{
		ClientId:        r.ClientId,
		Secret:          r.Secret,
		ApplicationType: r.ApplicationType,
		Grants:          r.Grants,
		RedirectURIs:    r.RedirectURIs,
		Scopes:          r.Scopes,
	}
	m.ID = r.ID
	return
}
