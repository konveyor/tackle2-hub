package resource

import (
	"github.com/konveyor/tackle2-hub/internal/assessment"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Thresholds REST resource.
type Thresholds api.Thresholds

// RiskMessages REST resource.
type RiskMessages api.RiskMessages

// CategorizedTag REST resource.
type CategorizedTag api.CategorizedTag

// Assessment REST resource.
type Assessment api.Assessment

// With updates the resource with the model.
func (r *Assessment) With(m *model.Assessment) {
	baseWith(&r.Resource, &m.Model)
	r.Questionnaire = ref(m.QuestionnaireID, &m.Questionnaire)
	r.Archetype = refPtr(m.ArchetypeID, m.Archetype)
	r.Application = refPtr(m.ApplicationID, m.Application)
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, Ref{
			ID:   s.ID,
			Name: s.Name,
		})
	}
	r.StakeholderGroups = []Ref{}
	for _, sg := range m.StakeholderGroups {
		r.StakeholderGroups = append(r.StakeholderGroups, Ref{
			ID:   sg.ID,
			Name: sg.Name,
		})
	}
	a := assessment.Assessment{}
	a.With(m)
	r.Required = a.Questionnaire.Required
	r.Risk = a.Risk()
	r.Confidence = a.Confidence()
	r.RiskMessages = api.RiskMessages(a.RiskMessages)
	r.Thresholds = api.Thresholds(a.Thresholds)
	r.Sections = []api.Section{}
	for _, s := range a.Sections {
		sect := Section{}
		sect.With(&s)
		r.Sections = append(r.Sections, api.Section(sect))
	}
	r.Status = a.Status()
}

// Model builds a model.
func (r *Assessment) Model() (m *model.Assessment) {
	m = &model.Assessment{}
	m.ID = r.ID
	for _, s := range r.Sections {
		sect := Section(s)
		m.Sections = append(m.Sections, *sect.Model())
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

// Question REST resource.
type Question api.Question

// With updates the resource with the model.
func (r *Question) With(m *model.Question) {
	r.Order = m.Order
	r.Text = m.Text
	r.Explanation = m.Explanation
	r.IncludeFor = []api.CategorizedTag{}
	for _, t := range m.IncludeFor {
		r.IncludeFor = append(r.IncludeFor, api.CategorizedTag(t))
	}
	r.ExcludeFor = []api.CategorizedTag{}
	for _, t := range m.ExcludeFor {
		r.ExcludeFor = append(r.ExcludeFor, api.CategorizedTag(t))
	}
	r.Answers = []api.Answer{}
	for _, a := range m.Answers {
		answer := Answer{}
		answer.With(&a)
		r.Answers = append(r.Answers, api.Answer(answer))
	}
}

// Model builds a model.
func (r *Question) Model() (m *model.Question) {
	m = &model.Question{
		Order:       r.Order,
		Text:        r.Text,
		Explanation: r.Explanation,
	}
	for _, t := range r.IncludeFor {
		m.IncludeFor = append(m.IncludeFor, model.CategorizedTag(t))
	}
	for _, t := range r.ExcludeFor {
		m.ExcludeFor = append(m.ExcludeFor, model.CategorizedTag(t))
	}
	for _, a := range r.Answers {
		answer := Answer(a)
		m.Answers = append(m.Answers, *answer.Model())
	}
	return
}

// Answer REST resource.
type Answer api.Answer

// With updates the resource with the model.
func (r *Answer) With(m *model.Answer) {
	r.Order = m.Order
	r.Text = m.Text
	r.Risk = m.Risk
	r.Rationale = m.Rationale
	r.Mitigation = m.Mitigation
	r.Selected = m.Selected
	r.AutoAnswered = m.AutoAnswered
	r.ApplyTags = []api.CategorizedTag{}
	for _, t := range m.ApplyTags {
		r.ApplyTags = append(r.ApplyTags, api.CategorizedTag(t))
	}
	r.AutoAnswerFor = []api.CategorizedTag{}
	for _, t := range m.AutoAnswerFor {
		r.AutoAnswerFor = append(r.AutoAnswerFor, api.CategorizedTag(t))
	}
}

// Model builds a model.
func (r *Answer) Model() (m *model.Answer) {
	m = &model.Answer{
		Order:        r.Order,
		Text:         r.Text,
		Risk:         r.Risk,
		Rationale:    r.Rationale,
		Mitigation:   r.Mitigation,
		Selected:     r.Selected,
		AutoAnswered: r.AutoAnswered,
	}
	for _, t := range r.ApplyTags {
		m.ApplyTags = append(m.ApplyTags, model.CategorizedTag(t))
	}
	for _, t := range r.AutoAnswerFor {
		m.AutoAnswerFor = append(m.AutoAnswerFor, model.CategorizedTag(t))
	}
	return
}

// Section REST resource.
type Section api.Section

// With updates the resource with the model.
func (r *Section) With(m *model.Section) {
	r.Order = m.Order
	r.Name = m.Name
	r.Comment = m.Comment
	r.Questions = []api.Question{}
	for _, q := range m.Questions {
		question := Question{}
		question.With(&q)
		r.Questions = append(r.Questions, api.Question(question))
	}
}

// Model builds a model.
func (r *Section) Model() (m *model.Section) {
	m = &model.Section{
		Order:   r.Order,
		Name:    r.Name,
		Comment: r.Comment,
	}
	for _, q := range r.Questions {
		question := Question(q)
		m.Questions = append(m.Questions, *question.Model())
	}
	return
}
