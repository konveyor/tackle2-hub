package assessment

import (
	"github.com/konveyor/tackle2-hub/internal/model"
)

// Archetype represents an Archetype with its assessments.
type Archetype struct {
	*model.Archetype
	Assessments []Assessment
}

// With updates the Archetype with the db model and deserializes its assessments.
func (r *Archetype) With(m *model.Archetype) {
	r.Archetype = m
	for i := range m.Assessments {
		a := &m.Assessments[i]
		assessment := Assessment{}
		assessment.With(a)
		r.Assessments = append(r.Assessments, assessment)
	}
}

// NewArchetypeResolver creates a new ArchetypeResolver.
func NewArchetypeResolver(
	m *model.Archetype,
	tags *TagResolver,
	membership *MembershipResolver,
	questionnaire *QuestionnaireResolver) (a *ArchetypeResolver) {
	a = &ArchetypeResolver{
		tags:          tags,
		membership:    membership,
		questionnaire: questionnaire,
	}
	archetype := Archetype{}
	archetype.With(m)
	a.archetype = archetype
	return
}

// ArchetypeResolver wraps an Archetype model
// with assessment-related functionality.
type ArchetypeResolver struct {
	archetype     Archetype
	tags          *TagResolver
	membership    *MembershipResolver
	questionnaire *QuestionnaireResolver
}

// AssessmentTags returns the list of tags that the archetype should
// inherit from the answers given to its assessments.
func (r *ArchetypeResolver) AssessmentTags() (tags []model.Tag) {
	if r.tags == nil {
		return
	}
	seenTags := make(map[uint]bool)
	for _, assessment := range r.archetype.Assessments {
		aTags := r.tags.Assessment(assessment)
		for _, t := range aTags {
			if _, found := seenTags[t.ID]; !found {
				seenTags[t.ID] = true
				tags = append(tags, t)
			}
		}
	}
	return
}

// RequiredAssessments returns the slice of assessments that are for required questionnaires.
func (r *ArchetypeResolver) RequiredAssessments() (required []Assessment) {
	for _, a := range r.archetype.Assessments {
		if r.questionnaire.Required(a.QuestionnaireID) {
			required = append(required, a)
		}
	}
	return
}

// Risk returns the overall risk level for the archetypes' assessments.
func (r *ArchetypeResolver) Risk() (risk string) {
	risk = Risk(r.RequiredAssessments())
	return
}

// Confidence returns the archetype's overall assessment confidence score.
func (r *ArchetypeResolver) Confidence() (confidence int) {
	confidence = Confidence(r.RequiredAssessments())
	return
}

// Assessed returns whether the archetype has been fully assessed.
func (r *ArchetypeResolver) Assessed() (assessed bool) {
	if r.questionnaire == nil {
		return
	}
	assessed = r.questionnaire.Assessed(r.RequiredAssessments())
	return
}

// Applications returns the archetype's member applications.
func (r *ArchetypeResolver) Applications() (applications []Application, err error) {
	if r.membership == nil {
		return
	}
	applications, err = r.membership.Applications(r.archetype)
	return
}
