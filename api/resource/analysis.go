package resource

import (
	"sort"

	"github.com/konveyor/tackle2-hub/model"
)

// Analysis REST resource.
type Analysis struct {
	Resource     `yaml:",inline"`
	Application  Ref               `json:"application"`
	Effort       int               `json:"effort"`
	Commit       string            `json:"commit,omitempty" yaml:",omitempty"`
	Archived     bool              `json:"archived,omitempty" yaml:",omitempty"`
	Insights     []Insight         `json:"insights,omitempty" yaml:",omitempty"`
	Dependencies []TechDependency  `json:"dependencies,omitempty" yaml:",omitempty"`
	Summary      []ArchivedInsight `json:"summary,omitempty" yaml:",omitempty" swaggertype:"object"`
}

// With updates the resource with the model.
func (r *Analysis) With(m *model.Analysis) {
	r.Resource.With(&m.Model)
	r.Application = r.ref(m.ApplicationID, m.Application)
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
type Insight struct {
	Resource    `yaml:",inline"`
	Analysis    uint       `json:"analysis"`
	RuleSet     string     `json:"ruleset" binding:"required"`
	Rule        string     `json:"rule" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty" yaml:",omitempty"`
	Category    string     `json:"category,omitempty" yaml:",omitempty"`
	Effort      int        `json:"effort,omitempty" yaml:",omitempty"`
	Incidents   []Incident `json:"incidents,omitempty" yaml:",omitempty"`
	Links       []Link     `json:"links,omitempty" yaml:",omitempty"`
	Facts       Map        `json:"facts,omitempty" yaml:",omitempty"`
	Labels      []string   `json:"labels"`
}

// With updates the resource with the model.
func (r *Insight) With(m *model.Insight) {
	r.Resource.With(&m.Model)
	r.Analysis = m.AnalysisID
	r.RuleSet = m.RuleSet
	r.Rule = m.Rule
	r.Name = m.Name
	r.Description = m.Description
	r.Category = m.Category
	r.Incidents = []Incident{}
	for i := range m.Incidents {
		n := Incident{}
		n.With(&m.Incidents[i])
		r.Incidents = append(
			r.Incidents,
			n)
	}
	for _, l := range m.Links {
		r.Links = append(r.Links, Link(l))
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
		n := r.Incidents[i].Model()
		m.Incidents = append(m.Incidents, *n)
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
type TechDependency struct {
	Resource `yaml:",inline"`
	Analysis uint     `json:"analysis"`
	Provider string   `json:"provider" yaml:",omitempty"`
	Name     string   `json:"name" binding:"required"`
	Version  string   `json:"version,omitempty" yaml:",omitempty"`
	Indirect bool     `json:"indirect,omitempty" yaml:",omitempty"`
	Labels   []string `json:"labels,omitempty" yaml:",omitempty"`
	SHA      string   `json:"sha,omitempty" yaml:",omitempty"`
}

// With updates the resource with the model.
func (r *TechDependency) With(m *model.TechDependency) {
	r.Resource.With(&m.Model)
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
type Incident struct {
	Resource `yaml:",inline"`
	Insight  uint   `json:"insight"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	CodeSnip string `json:"codeSnip" yaml:"codeSnip"`
	Facts    Map    `json:"facts"`
}

// With updates the resource with the model.
func (r *Incident) With(m *model.Incident) {
	r.Resource.With(&m.Model)
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

// Link analysis report link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty" yaml:",omitempty"`
}

// ArchivedInsight created when insights are archived.
type ArchivedInsight model.ArchivedInsight
