package resource

import "github.com/konveyor/tackle2-hub/model"

// AnalysisProfile REST resource.
type AnalysisProfile struct {
	Resource    `yaml:",inline"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty" yaml:",omitempty"`
	Mode        ApMode  `json:"mode"`
	Scope       ApScope `json:"scope"`
	Rules       ApRules `json:"rules"`
}

// With updates the resource with the model.
func (r *AnalysisProfile) With(m *model.AnalysisProfile) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Mode.WithDeps = m.WithDeps
	r.Scope.WithKnownLibs = m.WithKnownLibs
	r.Scope.Packages = m.Packages
	r.Rules.Labels = m.Labels
	if m.Repository != (model.Repository{}) {
		repository := Repository(m.Repository)
		r.Rules.Repository = &repository
	}
	r.Rules.Targets = make([]Ref, len(m.Targets))
	for i, t := range m.Targets {
		r.Rules.Targets[i] =
			Ref{
				ID:   t.ID,
				Name: t.Name,
			}
	}
	r.Rules.Files = make([]Ref, len(m.Files))
	for i, f := range m.Files {
		r.Rules.Files[i] = Ref(f)
	}
}

// Model builds a model.
func (r *AnalysisProfile) Model() (m *model.AnalysisProfile) {
	m = &model.AnalysisProfile{}
	m.Name = r.Name
	m.Description = r.Description
	m.WithDeps = r.Mode.WithDeps
	m.WithKnownLibs = r.Scope.WithKnownLibs
	m.Packages = r.Scope.Packages
	m.Labels = r.Rules.Labels
	if r.Rules.Repository != nil {
		m.Repository = model.Repository(*r.Rules.Repository)
	}
	m.Targets = make([]model.Target, len(r.Rules.Targets))
	for i, t := range r.Rules.Targets {
		m.Targets[i] =
			model.Target{
				Model: model.Model{
					ID: t.ID,
				},
			}
	}
	m.Files = make([]model.Ref, len(r.Rules.Files))
	for i, f := range r.Rules.Files {
		m.Files[i] = model.Ref(f)
	}
	m.ID = r.ID
	return
}

// ApMode analysis mode.
type ApMode struct {
	WithDeps bool `json:"withDeps" yaml:"withDeps"`
}

// ApScope analysis scope.
type ApScope struct {
	WithKnownLibs bool     `json:"withKnownLibs" yaml:"withKnownLibs"`
	Packages      InExList `json:"packages,omitempty" yaml:",omitempty"`
}

// ApRules analysis rules.
type ApRules struct {
	Targets    []Ref       `json:"targets"`
	Labels     InExList    `json:"labels,omitempty" yaml:",omitempty"`
	Files      []Ref       `json:"files,omitempty" yaml:",omitempty"`
	Repository *Repository `json:"repository,omitempty" yaml:",omitempty"`
}
