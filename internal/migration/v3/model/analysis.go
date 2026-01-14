package model

// RuleBundle - Analysis rules.
type RuleBundle struct {
	Model
	Kind        string
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Custom      bool
	Repository  JSON `gorm:"type:json"`
	ImageID     uint `gorm:"index" ref:"file"`
	Image       *File
	IdentityID  *uint `gorm:"index"`
	Identity    *Identity
	RuleSets    []RuleSet `gorm:"constraint:OnDelete:CASCADE"`
}

// RuleSet - Analysis ruleset.
type RuleSet struct {
	Model
	Name         string `gorm:"uniqueIndex:RuleSetA;not null"`
	Description  string
	Metadata     JSON `gorm:"type:json"`
	RuleBundleID uint `gorm:"uniqueIndex:RuleSetA;not null"`
	RuleBundle   *RuleBundle
	FileID       *uint `gorm:"index" ref:"file"`
	File         *File
}
