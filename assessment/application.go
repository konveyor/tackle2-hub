package assessment

import (
	"github.com/konveyor/tackle2-hub/model"
)

//
// NewApplicationResolver creates a new ApplicationResolver from an application and other shared resolvers.
func NewApplicationResolver(app *model.Application, tags *TagResolver, membership *MembershipResolver, questionnaire *QuestionnaireResolver) (a *ApplicationResolver) {
	a = &ApplicationResolver{
		application:           app,
		tagResolver:           tags,
		membershipResolver:    membership,
		questionnaireResolver: questionnaire,
	}
	return
}

//
// ApplicationResolver wraps an Application model
// with archetype and assessment resolution behavior.
type ApplicationResolver struct {
	application           *model.Application
	archetypes            []model.Archetype
	tagResolver           *TagResolver
	membershipResolver    *MembershipResolver
	questionnaireResolver *QuestionnaireResolver
}

//
// Archetypes returns the list of archetypes the application is a member of.
func (r *ApplicationResolver) Archetypes() (archetypes []model.Archetype, err error) {
	if len(r.archetypes) > 0 {
		archetypes = r.archetypes
		return
	}

	archetypes, err = r.membershipResolver.Archetypes(r.application)
	return
}

//
// ArchetypeTags returns the list of tags that the application should inherit from the archetypes it is a member of,
// including any tags that would be inherited due to answers given to the archetypes' assessments.
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
		// if an application has any of its own assessments then it should not
		// inherit assessment tags from any of its archetypes.
		if len(r.application.Assessments) == 0 {
			for _, assessment := range a.Assessments {
				aTags := r.tagResolver.Assessment(&assessment)
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

//
// AssessmentTags returns the list of tags that the application should inherit from the answers given
// to its assessments.
func (r *ApplicationResolver) AssessmentTags() (tags []model.Tag) {
	seenTags := make(map[uint]bool)
	for _, assessment := range r.application.Assessments {
		aTags := r.tagResolver.Assessment(&assessment)
		for _, t := range aTags {
			if _, found := seenTags[t.ID]; !found {
				seenTags[t.ID] = true
				tags = append(tags, t)
			}
		}
	}
	return
}

//
// Risk returns the overall risk level for the application based on its or its archetypes' assessments.
func (r *ApplicationResolver) Risk() (risk string, err error) {
	var assessments []model.Assessment
	if len(r.application.Assessments) > 0 {
		assessments = r.application.Assessments
	} else {
		archetypes, aErr := r.Archetypes()
		if aErr != nil {
			err = aErr
			return
		}
		for _, a := range archetypes {
			assessments = append(assessments, a.Assessments...)
		}
	}
	risk = r.questionnaireResolver.Risk(r.application.Assessments)

	return
}

//
// Confidence returns the application's overall assessment confidence score.
func (r *ApplicationResolver) Confidence() (confidence int, err error) {
	var assessments []model.Assessment
	if len(r.application.Assessments) > 0 {
		assessments = r.application.Assessments
	} else {
		archetypes, aErr := r.Archetypes()
		if aErr != nil {
			err = aErr
			return
		}
		for _, a := range archetypes {
			assessments = append(assessments, a.Assessments...)
		}
	}
	confidence = r.questionnaireResolver.Confidence(r.application.Assessments)

	return
}

//
// Assessed returns whether the application has been fully assessed.
func (r *ApplicationResolver) Assessed() (assessed bool, err error) {
	// if the application has any of its own assessments, only consider them for
	// determining whether it has been assessed.
	if len(r.application.Assessments) > 0 {
		assessed = r.questionnaireResolver.Assessed(r.application.Assessments)
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
