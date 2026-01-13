package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Proxy REST resource.
type Proxy api.Proxy

// With updates the resource with the model.
func (r *Proxy) With(m *model.Proxy) {
	baseWith(&r.Resource, &m.Model)
	r.Enabled = m.Enabled
	r.Kind = m.Kind
	r.Host = m.Host
	r.Port = m.Port
	r.Identity = refPtr(m.IdentityID, m.Identity)
	r.Excluded = m.Excluded
	if r.Excluded == nil {
		r.Excluded = []string{}
	}
}

// Model builds a model.
func (r *Proxy) Model() (m *model.Proxy) {
	m = &model.Proxy{
		Enabled: r.Enabled,
		Kind:    r.Kind,
		Host:    r.Host,
		Port:    r.Port,
	}
	m.ID = r.ID
	m.IdentityID = idPtr(r.Identity)
	m.Excluded = r.Excluded

	return
}
