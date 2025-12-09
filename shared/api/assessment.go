package api

// Assessment REST resource.
type Assessment struct {
	Resource          `yaml:",inline"`
	Application       *Ref         `json:"application,omitempty" yaml:",omitempty" binding:"excluded_with=Archetype"`
	Archetype         *Ref         `json:"archetype,omitempty" yaml:",omitempty" binding:"excluded_with=Application"`
	Questionnaire     Ref          `json:"questionnaire" binding:"required"`
	Sections          []Section    `json:"sections" binding:"dive"`
	Stakeholders      []Ref        `json:"stakeholders"`
	StakeholderGroups []Ref        `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	Risk              string       `json:"risk"`
	Confidence        int          `json:"confidence"`
	Status            string       `json:"status"`
	Thresholds        Thresholds   `json:"thresholds"`
	RiskMessages      RiskMessages `json:"riskMessages" yaml:"riskMessages"`
	Required          bool         `json:"required"`
}

// Section assessment section.
type Section struct {
	Order     uint       `json:"order" yaml:"order"`
	Name      string     `json:"name" yaml:"name"`
	Questions []Question `json:"questions" yaml:"questions" binding:"min=1,dive"`
	Comment   string     `json:"comment,omitempty" yaml:"comment,omitempty"`
}

// Question represents a question in a questionnaire.
type Question struct {
	Order       uint             `json:"order" yaml:"order"`
	Text        string           `json:"text" yaml:"text"`
	Explanation string           `json:"explanation" yaml:"explanation"`
	IncludeFor  []CategorizedTag `json:"includeFor,omitempty" yaml:"includeFor,omitempty"`
	ExcludeFor  []CategorizedTag `json:"excludeFor,omitempty" yaml:"excludeFor,omitempty"`
	Answers     []Answer         `json:"answers" yaml:"answers" binding:"min=1,dive"`
}

// Answer represents an answer to a question in a questionnaire.
type Answer struct {
	Order         uint             `json:"order" yaml:"order"`
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

// Thresholds assessment thresholds.
type Thresholds struct {
	Red     uint `json:"red" yaml:"red"`
	Yellow  uint `json:"yellow" yaml:"yellow"`
	Unknown uint `json:"unknown" yaml:"unknown"`
}

// RiskMessages assessment risk messages.
type RiskMessages struct {
	Red     string `json:"red" yaml:"red"`
	Yellow  string `json:"yellow" yaml:"yellow"`
	Green   string `json:"green" yaml:"green"`
	Unknown string `json:"unknown" yaml:"unknown"`
}

type Questionnaire struct {
	Resource     `yaml:",inline"`
	Name         string       `json:"name" yaml:"name" binding:"required"`
	Description  string       `json:"description" yaml:"description"`
	Required     bool         `json:"required" yaml:"required"`
	Sections     []Section    `json:"sections" yaml:"sections" binding:"required,min=1,dive"`
	Thresholds   Thresholds   `json:"thresholds" yaml:"thresholds" binding:"required"`
	RiskMessages RiskMessages `json:"riskMessages" yaml:"riskMessages" binding:"required"`
	Builtin      bool         `json:"builtin,omitempty" yaml:"builtin,omitempty"`
}
