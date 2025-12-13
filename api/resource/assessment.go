package resource

import (
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
)

// Assessment REST resource.
type Assessment struct {
	Resource          `yaml:",inline"`
	Application       *Ref      `json:"application,omitempty" yaml:",omitempty" binding:"excluded_with=Archetype"`
	Archetype         *Ref      `json:"archetype,omitempty" yaml:",omitempty" binding:"excluded_with=Application"`
	Questionnaire     Ref       `json:"questionnaire" binding:"required"`
	Sections          []Section `json:"sections" binding:"dive"`
	Stakeholders      []Ref     `json:"stakeholders"`
	StakeholderGroups []Ref     `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	// read only
	Risk         string       `json:"risk"`
	Confidence   int          `json:"confidence"`
	Status       string       `json:"status"`
	Thresholds   Thresholds   `json:"thresholds"`
	RiskMessages RiskMessages `json:"riskMessages" yaml:"riskMessages"`
	Required     bool         `json:"required"`
}

type Section model.Section
type Thresholds model.Thresholds
type RiskMessages model.RiskMessages

// With updates the resource with the model.
func (r *Assessment) With(m *model.Assessment) {
	r.Resource.With(&m.Model)
	r.Questionnaire = r.ref(m.QuestionnaireID, &m.Questionnaire)
	r.Archetype = r.refPtr(m.ArchetypeID, m.Archetype)
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
	r.StakeholderGroups = []Ref{}
	for _, sg := range m.StakeholderGroups {
		ref := Ref{}
		ref.With(sg.ID, sg.Name)
		r.StakeholderGroups = append(r.StakeholderGroups, ref)
	}
	a := assessment.Assessment{}
	a.With(m)
	r.Required = a.Questionnaire.Required
	r.Risk = a.Risk()
	r.Confidence = a.Confidence()
	r.RiskMessages = RiskMessages(a.RiskMessages)
	r.Thresholds = Thresholds(a.Thresholds)
	r.Sections = []Section{}
	for _, s := range a.Sections {
		r.Sections = append(r.Sections, Section(s))
	}
	r.Status = a.Status()
}

// Model builds a model.
func (r *Assessment) Model() (m *model.Assessment) {
	m = &model.Assessment{}
	m.ID = r.ID
	for _, s := range r.Sections {
		m.Sections = append(m.Sections, model.Section(s))
	}
	m.QuestionnaireID = r.Questionnaire.ID
	if r.Archetype != nil {
		m.ArchetypeID = &r.Archetype.ID
	}
	if r.Application != nil {
		m.ApplicationID = &r.Application.ID
	}
	for _, ref := range r.Stakeholders {
		m.Stakeholders = append(
			m.Stakeholders,
			model.Stakeholder{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.StakeholderGroups {
		m.StakeholderGroups = append(
			m.StakeholderGroups,
			model.StakeholderGroup{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}
