package resource

import "github.com/konveyor/tackle2-hub/model"

// RuleSet REST resource.
type RuleSet struct {
	Resource    `yaml:",inline"`
	Kind        string      `json:"kind,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Rules       []Rule      `json:"rules"`
	Repository  *Repository `json:"repository,omitempty"`
	Identity    *Ref        `json:"identity,omitempty"`
	DependsOn   []Ref       `json:"dependsOn" yaml:"dependsOn"`
}

// With updates the resource with the model.
func (r *RuleSet) With(m *model.RuleSet) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Description = m.Description
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	if m.Repository != (model.Repository{}) {
		repo := Repository(m.Repository)
		r.Repository = &repo
	}
	r.Rules = []Rule{}
	for i := range m.Rules {
		rule := Rule{}
		rule.With(&m.Rules[i])
		r.Rules = append(
			r.Rules,
			rule)
	}
	r.DependsOn = []Ref{}
	for i := range m.DependsOn {
		dep := Ref{}
		dep.With(m.DependsOn[i].ID, m.DependsOn[i].Name)
		r.DependsOn = append(r.DependsOn, dep)
	}
}

// Model builds a model.
func (r *RuleSet) Model() (m *model.RuleSet) {
	m = &model.RuleSet{
		Kind:        r.Kind,
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	m.IdentityID = r.idPtr(r.Identity)
	m.Rules = []model.Rule{}
	for _, rule := range r.Rules {
		m.Rules = append(m.Rules, *rule.Model())
	}
	if r.Repository != nil {
		m.Repository = model.Repository(*r.Repository)
	}
	for _, ref := range r.DependsOn {
		m.DependsOn = append(
			m.DependsOn,
			model.RuleSet{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	return
}

// Rule - REST Resource.
type Rule struct {
	Resource    `yaml:",inline"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	File        *Ref     `json:"file,omitempty"`
}

// With updates the resource with the model.
func (r *Rule) With(m *model.Rule) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Labels = m.Labels
	r.File = r.refPtr(m.FileID, m.File)
}

// Model builds a model.
func (r *Rule) Model() (m *model.Rule) {
	m = &model.Rule{}
	m.ID = r.ID
	m.Name = r.Name
	m.Labels = r.Labels
	m.FileID = r.idPtr(r.File)
	return
}
