package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// RuleSet REST resource.
type RuleSet api.RuleSet

// With updates the resource with the model.
func (r *RuleSet) With(m *model.RuleSet) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Description = m.Description
	r.Identity = refPtr(m.IdentityID, m.Identity)
	if m.Repository != (model.Repository{}) {
		repo := Repository(m.Repository)
		r.Repository = &repo
	}
	r.Rules = []api.Rule{}
	for i := range m.Rules {
		rule := Rule{}
		rule.With(&m.Rules[i])
		r.Rules = append(
			r.Rules,
			api.Rule(rule))
	}
	r.DependsOn = []Ref{}
	for i := range m.DependsOn {
		r.DependsOn = append(r.DependsOn, Ref{
			ID:   m.DependsOn[i].ID,
			Name: m.DependsOn[i].Name,
		})
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
	m.IdentityID = idPtr(r.Identity)
	m.Rules = []model.Rule{}
	for _, rule := range r.Rules {
		rr := Rule(rule)
		m.Rules = append(m.Rules, *rr.Model())
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

// HasRule - determine if the ruleset is referenced.
func (r *RuleSet) HasRule(id uint) (b bool) {
	for _, ruleset := range r.Rules {
		if id == ruleset.ID {
			b = true
			break
		}
	}
	return
}

// Rule REST resource.
type Rule api.Rule

// With updates the resource with the model.
func (r *Rule) With(m *model.Rule) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Labels = m.Labels
	r.File = refPtr(m.FileID, m.File)
}

// Model builds a model.
func (r *Rule) Model() (m *model.Rule) {
	m = &model.Rule{}
	m.ID = r.ID
	m.Name = r.Name
	m.Labels = r.Labels
	m.FileID = idPtr(r.File)
	return
}

// TargetLabel REST resource.
type TargetLabel = api.TargetLabel
