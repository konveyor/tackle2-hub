package resource

import "github.com/konveyor/tackle2-hub/model"

// Platform REST resource.
type Platform struct {
	Resource     `yaml:",inline"`
	Kind         string `json:"kind" binding:"required"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Identity     *Ref   `json:"identity,omitempty" yaml:",omitempty"`
	Applications []Ref  `json:"applications,omitempty" yaml:",omitempty"`
}

// With updates the resource with the model.
func (r *Platform) With(m *model.Platform) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.URL = m.URL
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	r.Applications = make([]Ref, 0, len(m.Applications))
	for _, a := range m.Applications {
		r.Applications = append(r.Applications, r.ref(a.ID, &a))
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
