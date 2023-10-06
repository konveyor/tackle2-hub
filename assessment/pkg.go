package assessment

import (
	"encoding/json"
	"github.com/konveyor/tackle2-hub/model"
	"math"
)

//
// Assessment risk
const (
	RiskUnknown = "unknown"
	RiskRed     = "red"
	RiskYellow  = "yellow"
	RiskGreen   = "green"
)

//
// Confidence adjustment
const (
	AdjusterRed    = 0.5
	AdjusterYellow = 0.98
)

//
// Confidence multiplier.
const (
	MultiplierRed    = 0.6
	MultiplierYellow = 0.95
)

//
// Risk weights
const (
	WeightRed     = 1
	WeightYellow  = 80
	WeightGreen   = 100
	WeightUnknown = 70
)

//
// Confidence calculates a confidence score based on the answers to an assessment's questions.
// The algorithm is a reimplementation of the calculation done by Pathfinder.
func Confidence(sections []Section) (score int) {
	totalQuestions := 0
	riskCounts := make(map[string]int)
	for _, s := range sections {
		for _, r := range s.Risks() {
			riskCounts[r]++
			totalQuestions++
		}
	}
	adjuster := 1.0
	if riskCounts[RiskRed] > 0 {
		adjuster = adjuster * math.Pow(AdjusterRed, float64(riskCounts[RiskRed]))
	}
	if riskCounts[RiskYellow] > 0 {
		adjuster = adjuster * math.Pow(AdjusterYellow, float64(riskCounts[RiskYellow]))
	}
	confidence := 0.0
	for i := 0; i < riskCounts[RiskRed]; i++ {
		confidence *= MultiplierRed
		confidence += WeightRed * adjuster
	}
	for i := 0; i < riskCounts[RiskYellow]; i++ {
		confidence *= MultiplierYellow
		confidence += WeightYellow * adjuster
	}
	confidence += float64(riskCounts[RiskGreen]) * WeightGreen * adjuster
	confidence += float64(riskCounts[RiskUnknown]) * WeightUnknown * adjuster

	maxConfidence := WeightGreen * totalQuestions
	score = int(confidence / float64(maxConfidence) * 100)

	return
}

//
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

//
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
