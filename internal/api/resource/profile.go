package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// AnalysisProfile REST resource.
type AnalysisProfile api.AnalysisProfile

// With updates the resource with the model.
func (r *AnalysisProfile) With(m *model.AnalysisProfile) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Mode.WithDeps = m.WithDeps
	r.Scope.WithKnownLibs = m.WithKnownLibs
	r.Scope.Packages = api.InExList(m.Packages)
	r.Rules.Labels = api.InExList(m.Labels)
	if m.Repository != (model.Repository{}) {
		repository := api.Repository(m.Repository)
		r.Rules.Repository = &repository
	}
	if m.Identity != nil {
		r.Rules.Identity = refPtr(m.IdentityID, m.Identity)
	}
	mp := make(map[uint]string)
	for _, ref := range m.Selections {
		mp[ref.ID] = ref.Label
	}
	r.Rules.Targets = make([]api.ApTargetRef, len(m.Targets))
	for i, t := range m.Targets {
		r.Rules.Targets[i] = api.ApTargetRef{
			ID:        t.ID,
			Name:      t.Name,
			Selection: mp[t.ID],
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
	m.Packages = model.InExList(r.Scope.Packages)
	m.Labels = model.InExList(r.Rules.Labels)
	if r.Rules.Repository != nil {
		m.Repository = model.Repository(*r.Rules.Repository)
	}
	m.Selections = make([]model.TargetSelection, 0)
	m.Targets = make([]model.Target, len(r.Rules.Targets))
	for i, t := range r.Rules.Targets {
		if t.Selection != "" {
			m.Selections = append(
				m.Selections,
				model.TargetSelection{
					ID:    t.ID,
					Label: t.Selection,
				})
		}
		m.Targets[i] =
			model.Target{
				Model: model.Model{
					ID: t.ID,
				},
			}
	}
	if r.Rules.Identity != nil {
		m.IdentityID = &r.Rules.Identity.ID
	}
	m.Files = make([]model.Ref, len(r.Rules.Files))
	for i, f := range r.Rules.Files {
		m.Files[i] = model.Ref(f)
	}
	m.ID = r.ID
	return
}

// ApMode REST resource.
type ApMode struct {
	WithDeps bool `json:"withDeps" yaml:"withDeps"`
}

// ApScope REST resource.
type ApScope struct {
	WithKnownLibs bool     `json:"withKnownLibs" yaml:"withKnownLibs"`
	Packages      InExList `json:"packages,omitempty" yaml:",omitempty"`
}

// ApRules REST resource.
type ApRules struct {
	Targets    []Ref       `json:"targets"`
	Labels     InExList    `json:"labels,omitempty" yaml:",omitempty"`
	Files      []Ref       `json:"files,omitempty" yaml:",omitempty"`
	Repository *Repository `json:"repository,omitempty" yaml:",omitempty"`
}
