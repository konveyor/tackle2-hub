package assessment

import (
	"github.com/konveyor/tackle2-hub/internal/model"
)

// Assessment risk
const (
	RiskUnassessed = "unassessed"
	RiskUnknown    = "unknown"
	RiskRed        = "red"
	RiskYellow     = "yellow"
	RiskGreen      = "green"
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
	// Return "unassessed" immediately if there are no assessments
	if len(assessments) == 0 {
		return RiskUnassessed
	}

	red, yellow, unknown, green := 0, 0, 0, 0

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
	tagSet := NewSet()
	for _, t := range application.Tags {
		tagSet.Add(t.ID)
	}
	a := Assessment{}
	a.With(assessment)
	a.Prepare(tagResolver, tagSet)
	return
}

// PrepareForArchetype prepares the sections of an assessment by including, excluding,
// or auto-answering questions based on a set of tags.
func PrepareForArchetype(tagResolver *TagResolver, archetype *model.Archetype, assessment *model.Assessment) {
	tagSet := NewSet()
	for _, t := range archetype.CriteriaTags {
		tagSet.Add(t.ID)
	}
	for _, t := range archetype.Tags {
		tagSet.Add(t.ID)
	}
	a := Assessment{}
	a.With(assessment)
	a.Prepare(tagResolver, tagSet)
	return
}
