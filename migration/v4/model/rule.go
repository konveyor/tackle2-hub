package model

//
// Rule - Analysis rule.
type Rule struct {
	Model
	Name        string `gorm:"uniqueIndex:ruleA;not null"`
	Description string
	RuleSetID   uint `gorm:"uniqueIndex:ruleA;not null"`
	RuleSet     *RuleSet
	FileID      *uint `gorm:"index" ref:"file"`
	File        *File
}
