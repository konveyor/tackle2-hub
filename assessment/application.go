package assessment

import (
	"github.com/konveyor/tackle2-hub/model"
)

// Application represents an Application with its assessments.
type Application struct {
	*model.Application
	Assessments []Assessment
}

// With updates the Application with the db model and deserializes its assessments.
func (r *Application) With(m *model.Application) {
	r.Application = m
	for i := range m.Assessments {
		a := &m.Assessments[i]
		assessment := Assessment{}
		assessment.With(a)
		r.Assessments = append(r.Assessments, assessment)
	}
}

// NewApplicationResolver creates a new ApplicationResolver from an application and other shared resolvers.
func NewApplicationResolver(m *model.Application, tags *TagResolver, membership *MembershipResolver, questionnaire *QuestionnaireResolver) (a *ApplicationResolver) {
	a = &ApplicationResolver{
		tagResolver:           tags,
		membershipResolver:    membership,
		questionnaireResolver: questionnaire,
	}
	app := Application{}
	app.With(m)
	a.application = app

	return
}

// ApplicationResolver wraps an Application model
// with archetype and assessment resolution behavior.
type ApplicationResolver struct {
	application           Application
	archetypes            []Archetype
	tagResolver           *TagResolver
	membershipResolver    *MembershipResolver
	questionnaireResolver *QuestionnaireResolver
}

// Archetypes returns the list of archetypes the application is a member of.
func (r *ApplicationResolver) Archetypes() (archetypes []Archetype, err error) {
	if len(r.archetypes) > 0 {
		archetypes = r.archetypes
		return
	}

	archetypes, err = r.membershipResolver.Archetypes(r.application)
	return
}

// ArchetypeTags returns the list of tags that the application should inherit from the archetypes it is a member of.
func (r *ApplicationResolver) ArchetypeTags() (tags []model.Tag, err error) {
	archetypes, err := r.Archetypes()
	if err != nil {
		return
	}

	seenTags := make(map[uint]bool)
	for _, a := range archetypes {
		for _, t := range a.Tags {
			if _, found := seenTags[t.ID]; !found {
				seenTags[t.ID] = true
				tags = append(tags, t)
			}
		}
	}
	return
}

// RequiredAssessments returns the slice of assessments that are for required questionnaires.
func (r *ApplicationResolver) RequiredAssessments() (required []Assessment) {
	for _, a := range r.application.Assessments {
		if r.questionnaireResolver.Required(a.QuestionnaireID) {
			required = append(required, a)
		}
	}
	return
}

// AssessmentTags returns the list of tags that the application should inherit from the answers given
// to its assessments or those of its archetypes. Archetype assessments are only inherited if the application
// does not have any answers to required questionnaires.
func (r *ApplicationResolver) AssessmentTags() (tags []model.Tag) {
	seenTags := make(map[uint]bool)
	if len(r.RequiredAssessments()) > 0 {
		for _, assessment := range r.RequiredAssessments() {
			aTags := r.tagResolver.Assessment(assessment)
			for _, t := range aTags {
				if _, found := seenTags[t.ID]; !found {
					seenTags[t.ID] = true
					tags = append(tags, t)
				}
			}
		}
		return
	}

	archetypes, err := r.Archetypes()
	if err != nil {
		return
	}
	for _, a := range archetypes {
		for _, assessment := range a.Assessments {
			if r.questionnaireResolver.Required(assessment.QuestionnaireID) {
				aTags := r.tagResolver.Assessment(assessment)
				for _, t := range aTags {
					if _, found := seenTags[t.ID]; !found {
						seenTags[t.ID] = true
						tags = append(tags, t)
					}
				}
			}
		}
	}
	return
}

// Risk returns the overall risk level for the application based on its or its archetypes' assessments.
func (r *ApplicationResolver) Risk() (risk string, err error) {
	var assessments []Assessment
	requiredAssessments := r.RequiredAssessments()
	if len(requiredAssessments) > 0 {
		assessments = requiredAssessments
	} else {
		archetypes, aErr := r.Archetypes()
		if aErr != nil {
			err = aErr
			return
		}
		for _, a := range archetypes {
			for _, assessment := range a.Assessments {
				if r.questionnaireResolver.Required(assessment.QuestionnaireID) {
					assessments = append(assessments, assessment)
				}
			}
		}
	}
	risk = Risk(assessments)
	return
}

// Confidence returns the application's overall assessment confidence score.
func (r *ApplicationResolver) Confidence() (confidence int, err error) {
	var assessments []Assessment
	requiredAssessments := r.RequiredAssessments()
	if len(requiredAssessments) > 0 {
		assessments = requiredAssessments
	} else {
		archetypes, aErr := r.Archetypes()
		if aErr != nil {
			err = aErr
			return
		}
		for _, a := range archetypes {
			for _, assessment := range a.Assessments {
				if r.questionnaireResolver.Required(assessment.QuestionnaireID) {
					assessments = append(assessments, assessment)
				}
			}
		}
	}
	confidence = Confidence(assessments)
	return
}

// Assessed returns whether the application has been fully assessed.
func (r *ApplicationResolver) Assessed() (assessed bool, err error) {
	// if the application has any of its own assessments, only consider them for
	// determining whether it has been assessed.
	assessments := r.RequiredAssessments()
	if len(assessments) > 0 {
		assessed = r.questionnaireResolver.Assessed(assessments)
		return
	}
	// otherwise the application is assessed if all of its archetypes are fully assessed.
	archetypes, err := r.Archetypes()
	if err != nil {
		return
	}
	assessedCount := 0
	for _, a := range archetypes {
		if r.questionnaireResolver.Assessed(a.Assessments) {
			assessedCount++
		}
	}
	assessed = assessedCount > 0 && assessedCount == len(archetypes)
	return
}
