package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Target REST resource.
type Target api.Target

// With updates the resource with the model.
func (r *Target) With(m *model.Target) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Provider = m.Provider
	r.Choice = m.Choice
	r.Custom = !m.Builtin()
	if m.RuleSet != nil {
		rs := RuleSet{}
		rs.With(m.RuleSet)
		r.RuleSet = (*api.RuleSet)(&rs)
	}
	imgRef := Ref{ID: m.ImageID}
	if m.Image != nil {
		imgRef.Name = m.Image.Name
	}
	r.Image = imgRef
	r.Labels = []api.TargetLabel{}
	for _, l := range m.Labels {
		r.Labels = append(r.Labels, api.TargetLabel(l))
	}
}

// Model builds a model.
func (r *Target) Model() (m *model.Target) {
	m = &model.Target{
		Name:        r.Name,
		Description: r.Description,
		Provider:    r.Provider,
		Choice:      r.Choice,
	}
	m.ID = r.ID
	m.ImageID = r.Image.ID
	for _, l := range r.Labels {
		m.Labels = append(m.Labels, model.TargetLabel(l))
	}
	return
}
