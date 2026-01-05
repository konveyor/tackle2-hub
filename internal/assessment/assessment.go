package assessment

import (
	"math"

	"github.com/konveyor/tackle2-hub/internal/model"
)

// Assessment represents a deserialized Assessment.
type Assessment struct {
	*model.Assessment
}

// With updates the Assessment with the db model and deserializes its fields.
func (r *Assessment) With(m *model.Assessment) {
	r.Assessment = m
}

// Status returns the started status of the assessment.
func (r *Assessment) Status() string {
	if r.Complete() {
		return StatusComplete
	} else if r.Started() {
		return StatusStarted
	} else {
		return StatusEmpty
	}
}

// Complete returns whether all sections have been completed.
func (r *Assessment) Complete() bool {
	for _, s := range r.Sections {
		if !r.sectionComplete(&s) {
			return false
		}
	}
	return true
}

// Started returns whether any sections have been started.
func (r *Assessment) Started() bool {
	for _, s := range r.Sections {
		if r.sectionStarted(&s) {
			return true
		}
	}
	return false
}

// Risk calculates the risk level (red, yellow, green, unknown) for the application.
func (r *Assessment) Risk() string {
	var total uint
	colors := make(map[string]uint)
	for _, s := range r.Sections {
		for _, risk := range r.sectionRisks(&s) {
			colors[risk]++
			total++
		}
	}
	if total == 0 {
		return RiskUnknown
	}
	if (float64(colors[RiskRed]) / float64(total)) >= (float64(r.Thresholds.Red) / float64(100)) {
		return RiskRed
	}
	if (float64(colors[RiskYellow]) / float64(total)) >= (float64(r.Thresholds.Yellow) / float64(100)) {
		return RiskYellow
	}
	if (float64(colors[RiskUnknown]) / float64(total)) >= (float64(r.Thresholds.Unknown) / float64(100)) {
		return RiskUnknown
	}
	return RiskGreen
}

// Confidence calculates a confidence score based on the answers to an assessment's questions.
// The algorithm is a reimplementation of the calculation done by Pathfinder.
func (r *Assessment) Confidence() (score int) {
	totalQuestions := 0
	riskCounts := make(map[string]int)
	for _, s := range r.Sections {
		for _, risk := range r.sectionRisks(&s) {
			riskCounts[risk]++
			totalQuestions++
		}
	}
	if totalQuestions == 0 {
		return
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

func (r *Assessment) Prepare(tagResolver *TagResolver, tags Set) {
	for i := range r.Sections {
		s := &r.Sections[i]
		includedQuestions := []model.Question{}
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
	return
}

func (r *Assessment) Tags() (tags []model.CategorizedTag) {
	for _, s := range r.Sections {
		for _, t := range r.sectionTags(&s) {
			tags = append(tags, t)
		}
	}
	return
}

// Complete returns whether all questions in the section have been answered.
func (r *Assessment) sectionComplete(s *model.Section) bool {
	for _, q := range s.Questions {
		if !r.questionAnswered(&q) {
			return false
		}
	}
	return true
}

// Started returns whether any questions in the section have been answered.

func (r *Assessment) sectionStarted(s *model.Section) bool {
	for _, q := range s.Questions {
		if r.questionAnswered(&q) && !r.questionAutoAnswered(&q) {
			return true
		}
	}
	return false
}

// Risks returns a slice of the risks of each of its questions.
func (r *Assessment) sectionRisks(s *model.Section) []string {
	risks := []string{}
	for _, q := range s.Questions {
		risks = append(risks, r.questionRisk(&q))
	}
	return risks
}

// Tags returns all the tags that should be applied based on how
// the questions in the section have been answered.
func (r *Assessment) sectionTags(s *model.Section) (tags []model.CategorizedTag) {
	for _, q := range s.Questions {
		tags = append(tags, r.questionTags(&q)...)
	}
	return
}

// Risk returns the risk level for the question based on how it has been answered.
func (r *Assessment) questionRisk(q *model.Question) string {
	for _, a := range q.Answers {
		if a.Selected {
			return a.Risk
		}
	}
	return RiskUnknown
}

// Answered returns whether the question has had an answer selected.
func (r *Assessment) questionAnswered(q *model.Question) bool {
	for _, a := range q.Answers {
		if a.Selected {
			return true
		}
	}
	return false
}

// AutoAnswered returns whether the question has had an
// answer pre-selected by the system.
func (r *Assessment) questionAutoAnswered(q *model.Question) bool {
	for _, a := range q.Answers {
		if a.AutoAnswered {
			return true
		}
	}
	return false
}

// Tags returns any tags to be applied based on how the question is answered.
func (r *Assessment) questionTags(q *model.Question) (tags []model.CategorizedTag) {
	for _, answer := range q.Answers {
		if answer.Selected {
			tags = answer.ApplyTags
			return
		}
	}
	return
}
