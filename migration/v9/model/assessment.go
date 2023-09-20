package model

type Questionnaire struct {
	Model
	UUID         *string `gorm:"uniqueIndex"`
	Name         string  `gorm:"unique"`
	Description  string
	Required     bool
	Sections     JSON         `gorm:"type:json"`
	Thresholds   JSON         `gorm:"type:json"`
	RiskMessages JSON         `gorm:"type:json"`
	Assessments  []Assessment `gorm:"constraint:OnDelete:CASCADE"`
}

type Assessment struct {
	Model
	ApplicationID     *uint `gorm:"uniqueIndex:AssessmentA"`
	Application       *Application
	ArchetypeID       *uint `gorm:"uniqueIndex:AssessmentB"`
	Archetype         *Archetype
	QuestionnaireID   uint `gorm:"uniqueIndex:AssessmentA;uniqueIndex:AssessmentB"`
	Questionnaire     Questionnaire
	Sections          JSON               `gorm:"type:json"`
	Thresholds        JSON               `gorm:"type:json"`
	RiskMessages      JSON               `gorm:"type:json"`
	Stakeholders      []Stakeholder      `gorm:"many2many:AssessmentStakeholders;constraint:OnDelete:CASCADE"`
	StakeholderGroups []StakeholderGroup `gorm:"many2many:AssessmentStakeholderGroups;constraint:OnDelete:CASCADE"`
}
