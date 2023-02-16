package model

//
// RuleBundle - Analysis rules.
type RuleBundle struct {
	Model
	Kind        string
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Custom      bool
	Repository  JSON
	ImageID     uint `gorm:"index" ref:"file"`
	Image       *File
	IdentityID  *uint `gorm:"index"`
	Identity    *Identity
	RuleSets    []RuleSet `gorm:"constraint:OnDelete:CASCADE"`
}
