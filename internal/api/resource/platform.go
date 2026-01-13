package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Platform REST resource.
type Platform api.Platform

// With updates the resource with the model.
func (r *Platform) With(m *model.Platform) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.URL = m.URL
	r.Identity = refPtr(m.IdentityID, m.Identity)
	r.Applications = make([]Ref, 0, len(m.Applications))
	for _, a := range m.Applications {
		r.Applications = append(r.Applications, ref(a.ID, &a))
	}
}

// Model builds a model.
func (r *Platform) Model() (m *model.Platform) {
	m = &model.Platform{}
	m.ID = r.ID
	m.Kind = r.Kind
	m.Name = r.Name
	m.URL = r.URL
	if r.Identity != nil {
		m.IdentityID = &r.Identity.ID
	}
	return
}
