package resource

import "github.com/konveyor/tackle2-hub/model"

// Target REST resource.
type Target struct {
	Resource    `yaml:",inline"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Provider    string        `json:"provider,omitempty" yaml:",omitempty"`
	Choice      bool          `json:"choice,omitempty" yaml:",omitempty"`
	Custom      bool          `json:"custom,omitempty" yaml:",omitempty"`
	Labels      []TargetLabel `json:"labels"`
	Image       Ref           `json:"image"`
	RuleSet     *RuleSet      `json:"ruleset,omitempty" yaml:"ruleset,omitempty"`
}

type TargetLabel model.TargetLabel

// With updates the resource with the model.
func (r *Target) With(m *model.Target) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Provider = m.Provider
	r.Choice = m.Choice
	r.Custom = !m.Builtin()
	if m.RuleSet != nil {
		r.RuleSet = &RuleSet{}
		r.RuleSet.With(m.RuleSet)
	}
	imgRef := Ref{ID: m.ImageID}
	if m.Image != nil {
		imgRef.Name = m.Image.Name
	}
	r.Image = imgRef
	r.Labels = []TargetLabel{}
	for _, l := range m.Labels {
		r.Labels = append(r.Labels, TargetLabel(l))
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
