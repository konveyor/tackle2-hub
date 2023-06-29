package model

//
// RuleSet - Analysis ruleset.
type RuleSet struct {
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
	Rules       []Rule    `gorm:"constraint:OnDelete:CASCADE"`
	DependsOn   []RuleSet `gorm:"many2many:RuleSetDependencies;constraint:OnDelete:CASCADE"`
}

//
// Rule - Analysis rule.
type Rule struct {
	Model
	Name        string
	Description string
	Labels      JSON `gorm:"type:json"`
	RuleSetID   uint `gorm:"uniqueIndex:RuleA;not null"`
	RuleSet     *RuleSet
	FileID      *uint `gorm:"uniqueIndex:RuleA" ref:"file"`
	File        *File
}
