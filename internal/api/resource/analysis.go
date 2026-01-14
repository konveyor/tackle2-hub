package resource

import (
	"sort"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Analysis REST resource.
type Analysis api.Analysis

// With updates the resource with the model.
func (r *Analysis) With(m *model.Analysis) {
	baseWith(&r.Resource, &m.Model)
	r.Application = ref(m.ApplicationID, m.Application)
	r.Effort = m.Effort
	r.Commit = m.Commit
	r.Archived = m.Archived
}

// Model builds a model.
func (r *Analysis) Model() (m *model.Analysis) {
	m = &model.Analysis{}
	m.Effort = r.Effort
	m.Commit = r.Commit
	m.Insights = []model.Insight{}
	return
}

// Insight REST resource.
type Insight api.Insight

// With updates the resource with the model.
func (r *Insight) With(m *model.Insight) {
	baseWith(&r.Resource, &m.Model)
	r.Analysis = m.AnalysisID
	r.RuleSet = m.RuleSet
	r.Rule = m.Rule
	r.Name = m.Name
	r.Description = m.Description
	r.Category = m.Category
	r.Incidents = []api.Incident{}
	for i := range m.Incidents {
		n := Incident{}
		n.With(&m.Incidents[i])
		r.Incidents = append(
			r.Incidents,
			api.Incident(n))
	}
	for _, l := range m.Links {
		r.Links = append(r.Links, api.Link(l))
	}
	r.Facts = m.Facts
	r.Labels = m.Labels
	r.Effort = m.Effort
}

// Model builds a model.
func (r *Insight) Model() (m *model.Insight) {
	m = &model.Insight{}
	m.RuleSet = r.RuleSet
	m.Rule = r.Rule
	m.Name = r.Name
	m.Description = r.Description
	m.Category = r.Category
	m.Incidents = []model.Incident{}
	for i := range r.Incidents {
		n := Incident(r.Incidents[i])
		mn := n.Model()
		m.Incidents = append(m.Incidents, *mn)
	}
	for _, l := range r.Links {
		m.Links = append(m.Links, model.Link(l))
	}
	m.Facts = r.Facts
	m.Labels = r.Labels
	m.Effort = r.Effort
	return
}

// TechDependency REST resource.
type TechDependency api.TechDependency

// With updates the resource with the model.
func (r *TechDependency) With(m *model.TechDependency) {
	baseWith(&r.Resource, &m.Model)
	r.Analysis = m.AnalysisID
	r.Provider = m.Provider
	r.Name = m.Name
	r.Version = m.Version
	r.Indirect = m.Indirect
	r.SHA = m.SHA
	r.Labels = m.Labels
}

// Model builds a model.
func (r *TechDependency) Model() (m *model.TechDependency) {
	sort.Strings(r.Labels)
	m = &model.TechDependency{}
	m.Name = r.Name
	m.Version = r.Version
	m.Provider = r.Provider
	m.Indirect = r.Indirect
	m.SHA = r.SHA
	m.Labels = r.Labels
	return
}

// Incident REST resource.
type Incident api.Incident

// With updates the resource with the model.
func (r *Incident) With(m *model.Incident) {
	baseWith(&r.Resource, &m.Model)
	r.Insight = m.InsightID
	r.File = m.File
	r.Line = m.Line
	r.Message = m.Message
	r.CodeSnip = m.CodeSnip
	r.Facts = m.Facts
}

// Model builds a model.
func (r *Incident) Model() (m *model.Incident) {
	m = &model.Incident{}
	m.File = r.File
	m.Line = r.Line
	m.Message = r.Message
	m.CodeSnip = r.CodeSnip
	m.Facts = r.Facts
	return
}

// Link REST resource.
type Link = api.Link

// ArchivedInsight REST resource.
type ArchivedInsight = api.ArchivedInsight

// DepAppReport REST resource.
type DepAppReport = api.DepAppReport

// DepReport REST resource.
type DepReport = api.DepReport

// FileReport REST resource.
type FileReport = api.FileReport

// InsightAppReport REST resource.
type InsightAppReport = api.InsightAppReport

// InsightReport REST resource.
type InsightReport = api.InsightReport

// RuleReport REST resource.
type RuleReport = api.RuleReport
