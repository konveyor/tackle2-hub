package model

//
// RuleSet - Analysis ruleset.
type RuleSet struct {
	Model
	Name         string `gorm:"uniqueIndex:RuleSetA;not null"`
	Description  string
	Metadata     JSON
	RuleBundleID uint `gorm:"uniqueIndex:RuleSetA;not null"`
	RuleBundle   *RuleBundle
	FileID       *uint `gorm:"index" ref:"file"`
	File         *File
}
