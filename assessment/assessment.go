package assessment

import (
	"encoding/json"
	"math"

	"github.com/konveyor/tackle2-hub/model"
)

// Assessment represents a deserialized Assessment.
type Assessment struct {
	*model.Assessment
	Sections     []Section    `json:"sections"`
	Thresholds   Thresholds   `json:"thresholds"`
	RiskMessages RiskMessages `json:"riskMessages"`
}

// With updates the Assessment with the db model and deserializes its fields.
func (r *Assessment) With(m *model.Assessment) {
	r.Assessment = m
	_ = json.Unmarshal(m.Sections, &r.Sections)
	_ = json.Unmarshal(m.Thresholds, &r.Thresholds)
	_ = json.Unmarshal(m.RiskMessages, &r.RiskMessages)
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
		if !s.Complete() {
			return false
		}
	}
	return true
}

// Started returns whether any sections have been started.
func (r *Assessment) Started() bool {
	for _, s := range r.Sections {
		if s.Started() {
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
		for _, risk := range s.Risks() {
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
		for _, r := range s.Risks() {
			riskCounts[r]++
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

// Section represents a group of questions in a questionnaire.
type Section struct {
	Order     *uint      `json:"order" yaml:"order" binding:"required"`
	Name      string     `json:"name" yaml:"name"`
	Questions []Question `json:"questions" yaml:"questions" binding:"min=1,dive"`
	Comment   string     `json:"comment,omitempty" yaml:"comment,omitempty"`
}

// Complete returns whether all questions in the section have been answered.
func (r *Section) Complete() bool {
	for _, q := range r.Questions {
		if !q.Answered() {
			return false
		}
	}
	return true
}

// Started returns whether any questions in the section have been answered.
func (r *Section) Started() bool {
	for _, q := range r.Questions {
		if q.Answered() {
			return true
		}
	}
	return false
}

// Risks returns a slice of the risks of each of its questions.
func (r *Section) Risks() []string {
	risks := []string{}
	for _, q := range r.Questions {
		risks = append(risks, q.Risk())
	}
	return risks
}

// Tags returns all the tags that should be applied based on how
// the questions in the section have been answered.
func (r *Section) Tags() (tags []CategorizedTag) {
	for _, q := range r.Questions {
		tags = append(tags, q.Tags()...)
	}
	return
}

// Question represents a question in a questionnaire.
type Question struct {
	Order       *uint            `json:"order" yaml:"order" binding:"required"`
	Text        string           `json:"text" yaml:"text"`
	Explanation string           `json:"explanation" yaml:"explanation"`
	IncludeFor  []CategorizedTag `json:"includeFor,omitempty" yaml:"includeFor,omitempty"`
	ExcludeFor  []CategorizedTag `json:"excludeFor,omitempty" yaml:"excludeFor,omitempty"`
	Answers     []Answer         `json:"answers" yaml:"answers" binding:"min=1,dive"`
}

// Risk returns the risk level for the question based on how it has been answered.
func (r *Question) Risk() string {
	for _, a := range r.Answers {
		if a.Selected {
			return a.Risk
		}
	}
	return RiskUnknown
}

// Answered returns whether the question has had an answer selected.
func (r *Question) Answered() bool {
	for _, a := range r.Answers {
		if a.Selected {
			return true
		}
	}
	return false
}

// Tags returns any tags to be applied based on how the question is answered.
func (r *Question) Tags() (tags []CategorizedTag) {
	for _, answer := range r.Answers {
		if answer.Selected {
			tags = answer.ApplyTags
			return
		}
	}
	return
}

// Answer represents an answer to a question in a questionnaire.
type Answer struct {
	Order         *uint            `json:"order" yaml:"order" binding:"required"`
	Text          string           `json:"text" yaml:"text"`
	Risk          string           `json:"risk" yaml:"risk" binding:"oneof=red yellow green unknown"`
	Rationale     string           `json:"rationale" yaml:"rationale"`
	Mitigation    string           `json:"mitigation" yaml:"mitigation"`
	ApplyTags     []CategorizedTag `json:"applyTags,omitempty" yaml:"applyTags,omitempty"`
	AutoAnswerFor []CategorizedTag `json:"autoAnswerFor,omitempty" yaml:"autoAnswerFor,omitempty"`
	Selected      bool             `json:"selected,omitempty" yaml:"selected,omitempty"`
	AutoAnswered  bool             `json:"autoAnswered,omitempty" yaml:"autoAnswered,omitempty"`
}

// CategorizedTag represents a human-readable pair of category and tag.
type CategorizedTag struct {
	Category string `json:"category" yaml:"category"`
	Tag      string `json:"tag" yaml:"tag"`
}

// RiskMessages contains messages to display for each risk level.
type RiskMessages struct {
	Red     string `json:"red" yaml:"red"`
	Yellow  string `json:"yellow" yaml:"yellow"`
	Green   string `json:"green" yaml:"green"`
	Unknown string `json:"unknown" yaml:"unknown"`
}

// Thresholds contains the threshold values for determining risk for the questionnaire.
type Thresholds struct {
	Red     uint `json:"red" yaml:"red"`
	Yellow  uint `json:"yellow" yaml:"yellow"`
	Unknown uint `json:"unknown" yaml:"unknown"`
}
