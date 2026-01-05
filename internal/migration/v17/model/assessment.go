package model

type Questionnaire struct {
	Model
	UUID         *string `gorm:"uniqueIndex"`
	Name         string  `gorm:"unique"`
	Description  string
	Required     bool
	Sections     []Section    `gorm:"type:json;serializer:json"`
	Thresholds   Thresholds   `gorm:"type:json;serializer:json"`
	RiskMessages RiskMessages `gorm:"type:json;serializer:json"`
	Assessments  []Assessment `gorm:"constraint:OnDelete:CASCADE"`
}

// Builtin returns true if this is a Konveyor-provided questionnaire.
func (r *Questionnaire) Builtin() bool {
	return r.UUID != nil
}

type Assessment struct {
	Model
	ApplicationID     *uint `gorm:"uniqueIndex:AssessmentA"`
	Application       *Application
	ArchetypeID       *uint `gorm:"uniqueIndex:AssessmentB"`
	Archetype         *Archetype
	QuestionnaireID   uint `gorm:"uniqueIndex:AssessmentA;uniqueIndex:AssessmentB"`
	Questionnaire     Questionnaire
	Sections          []Section          `gorm:"type:json;serializer:json"`
	Thresholds        Thresholds         `gorm:"type:json;serializer:json"`
	RiskMessages      RiskMessages       `gorm:"type:json;serializer:json"`
	Stakeholders      []Stakeholder      `gorm:"many2many:AssessmentStakeholders;constraint:OnDelete:CASCADE"`
	StakeholderGroups []StakeholderGroup `gorm:"many2many:AssessmentStakeholderGroups;constraint:OnDelete:CASCADE"`
}

type Review struct {
	Model
	BusinessCriticality uint   `gorm:"not null"`
	EffortEstimate      string `gorm:"not null"`
	ProposedAction      string `gorm:"not null"`
	WorkPriority        uint   `gorm:"not null"`
	Comments            string
	ApplicationID       *uint `gorm:"uniqueIndex"`
	Application         *Application
	ArchetypeID         *uint `gorm:"uniqueIndex"`
	Archetype           *Archetype
}

//
// JSON Fields.
//

// Section represents a group of questions in a questionnaire.
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
