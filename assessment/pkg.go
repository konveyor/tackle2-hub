package assessment

import (
	"encoding/json"
	"github.com/konveyor/tackle2-hub/model"
)

// Assessment risk
const (
	RiskUnknown = "unknown"
	RiskRed     = "red"
	RiskYellow  = "yellow"
	RiskGreen   = "green"
)

// Assessment status
const (
	StatusEmpty    = "empty"
	StatusStarted  = "started"
	StatusComplete = "complete"
)

// Confidence adjustment
const (
	AdjusterRed    = 0.5
	AdjusterYellow = 0.98
)

// Confidence multiplier.
const (
	MultiplierRed    = 0.6
	MultiplierYellow = 0.95
)

// Risk weights
const (
	WeightRed     = 1
	WeightYellow  = 80
	WeightGreen   = 100
	WeightUnknown = 70
)

// Risk returns the single highest risk score for a group of assessments.
func Risk(assessments []Assessment) (risk string) {
	risk = RiskUnknown
	if len(assessments) == 0 {
		return
	}
	red := 0
	yellow := 0
	unknown := 0
	green := 0
	if len(assessments) > 0 {
		for _, a := range assessments {
			switch a.Risk() {
			case RiskRed:
				red++
			case RiskYellow:
				yellow++
			case RiskGreen:
				green++
			default:
				unknown++
			}
		}
	}

	switch {
	case red > 0:
		risk = RiskRed
	case unknown > 0:
		risk = RiskUnknown
	case yellow > 0:
		risk = RiskYellow
	case green == len(assessments):
		risk = RiskGreen
	}

	return
}

// Confidence returns a total confidence score for a group of assessments.
func Confidence(assessments []Assessment) (confidence int) {
	if len(assessments) == 0 {
		return
	}
	for _, a := range assessments {
		confidence += a.Confidence()
	}
	confidence /= len(assessments)
	return
}

// PrepareForApplication prepares the sections of an assessment by including, excluding,
// or auto-answering questions based on a set of tags.
func PrepareForApplication(tagResolver *TagResolver, application *model.Application, assessment *model.Assessment) {
	sections := []Section{}
	_ = json.Unmarshal(assessment.Sections, &sections)

	tagSet := NewSet()
	for _, t := range application.Tags {
		tagSet.Add(t.ID)
	}

	assessment.Sections, _ = json.Marshal(prepareSections(tagResolver, tagSet, sections))

	return
}

// PrepareForArchetype prepares the sections of an assessment by including, excluding,
// or auto-answering questions based on a set of tags.
func PrepareForArchetype(tagResolver *TagResolver, archetype *model.Archetype, assessment *model.Assessment) {
	sections := []Section{}
	_ = json.Unmarshal(assessment.Sections, &sections)

	tagSet := NewSet()
	for _, t := range archetype.CriteriaTags {
		tagSet.Add(t.ID)
	}
	for _, t := range archetype.Tags {
		tagSet.Add(t.ID)
	}

	assessment.Sections, _ = json.Marshal(prepareSections(tagResolver, tagSet, sections))

	return
}

func prepareSections(tagResolver *TagResolver, tags Set, sections []Section) (preparedSections []Section) {
	for i := range sections {
		s := &sections[i]
		includedQuestions := []Question{}
		for _, q := range s.Questions {
			for j := range q.Answers {
				a := &q.Answers[j]
				autoAnswerTags := NewSet()
				for _, t := range a.AutoAnswerFor {
					tag, found := tagResolver.Resolve(t.Category, t.Tag)
					if found {
						autoAnswerTags.Add(tag.ID)
					}
				}
				if tags.Intersects(autoAnswerTags) {
					a.AutoAnswered = true
					a.Selected = true
					break
				}
			}

			if len(q.IncludeFor) > 0 {
				includeForTags := NewSet()
				for _, t := range q.IncludeFor {
					tag, found := tagResolver.Resolve(t.Category, t.Tag)
					if found {
						includeForTags.Add(tag.ID)
					}
				}
				if tags.Intersects(includeForTags) {
					includedQuestions = append(includedQuestions, q)
				}
				continue
			}

			if len(q.ExcludeFor) > 0 {
				excludeForTags := NewSet()
				for _, t := range q.ExcludeFor {
					tag, found := tagResolver.Resolve(t.Category, t.Tag)
					if found {
						excludeForTags.Add(tag.ID)
					}
				}
				if tags.Intersects(excludeForTags) {
					continue
				}
			}
			includedQuestions = append(includedQuestions, q)
		}
		s.Questions = includedQuestions
	}
	preparedSections = sections
	return
}
