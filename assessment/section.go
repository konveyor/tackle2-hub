package assessment

//
// Section represents a group of questions in a questionnaire.
type Section struct {
	Order     uint       `json:"order" yaml:"order" binding:"required"`
	Name      string     `json:"name" yaml:"name"`
	Questions []Question `json:"questions" yaml:"questions"`
	Comment   string     `json:"comment,omitempty" yaml:"comment,omitempty"`
}

//
// Complete returns whether all questions in the section have been answered.
func (r *Section) Complete() bool {
	for _, q := range r.Questions {
		if !q.Answered() {
			return false
		}
	}
	return true
}

//
// Started returns whether any questions in the section have been answered.
func (r *Section) Started() bool {
	for _, q := range r.Questions {
		if q.Answered() {
			return true
		}
	}
	return false
}

//
// Risks returns a slice of the risks of each of its questions.
func (r *Section) Risks() []string {
	risks := []string{}
	for _, q := range r.Questions {
		risks = append(risks, q.Risk())
	}
	return risks
}

//
// Tags returns all the tags that should be applied based on how
// the questions in the section have been answered.
func (r *Section) Tags() (tags []CategorizedTag) {
	for _, q := range r.Questions {
		tags = append(tags, q.Tags()...)
	}
	return
}

//
// Question represents a question in a questionnaire.
type Question struct {
	Order       uint             `json:"order" yaml:"order" binding:"required"`
	Text        string           `json:"text" yaml:"text"`
	Explanation string           `json:"explanation" yaml:"explanation"`
	IncludeFor  []CategorizedTag `json:"includeFor,omitempty" yaml:"includeFor,omitempty" binding:"excluded_with=ExcludeFor"`
	ExcludeFor  []CategorizedTag `json:"excludeFor,omitempty" yaml:"excludeFor,omitempty" binding:"excluded_with=IncludeFor"`
	Answers     []Answer         `json:"answers" yaml:"answers"`
}

//
// Risk returns the risk level for the question based on how it has been answered.
func (r *Question) Risk() string {
	for _, a := range r.Answers {
		if a.Selected {
			return a.Risk
		}
	}
	return RiskUnknown
}

//
// Answered returns whether the question has had an answer selected.
func (r *Question) Answered() bool {
	for _, a := range r.Answers {
		if a.Selected {
			return true
		}
	}
	return false
}

//
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

//
// Answer represents an answer to a question in a questionnaire.
type Answer struct {
	Order         uint             `json:"order" yaml:"order" binding:"required"`
	Text          string           `json:"text" yaml:"text"`
	Risk          string           `json:"risk" yaml:"risk" binding:"oneof=red,yellow,green,unknown"`
	Rationale     string           `json:"rationale" yaml:"rationale"`
	Mitigation    string           `json:"mitigation" yaml:"mitigation"`
	ApplyTags     []CategorizedTag `json:"applyTags,omitempty" yaml:"applyTags,omitempty"`
	AutoAnswerFor []CategorizedTag `json:"autoAnswerFor,omitempty" yaml:"autoAnswerFor,omitempty"`
	Selected      bool             `json:"selected,omitempty" yaml:"selected,omitempty"`
	AutoAnswered  bool             `json:"autoAnswered,omitempty" yaml:"autoAnswered,omitempty"`
}

//
// CategorizedTag represents a human-readable pair of category and tag.
type CategorizedTag struct {
	Category string `json:"category" yaml:"category"`
	Tag      string `json:"tag" yaml:"tag"`
}

//
// RiskMessages contains messages to display for each risk level.
type RiskMessages struct {
	Red     string `json:"red" yaml:"red"`
	Yellow  string `json:"yellow" yaml:"yellow"`
	Green   string `json:"green" yaml:"green"`
	Unknown string `json:"unknown" yaml:"unknown"`
}

//
// Thresholds contains the threshold values for determining risk for the questionnaire.
type Thresholds struct {
	Red     uint `json:"red" yaml:"red"`
	Yellow  uint `json:"yellow" yaml:"yellow"`
	Unknown uint `json:"unknown" yaml:"unknown"`
}
