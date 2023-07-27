package model

import "gorm.io/gorm"

//
// RuleSet - Analysis ruleset.
type RuleSet struct {
	Model
	UUID       *string `gorm:"uniqueIndex"`
	Kind       string
	Name       string `gorm:"uniqueIndex;not null"`
	Repository JSON   `gorm:"type:json"`
	IdentityID *uint  `gorm:"index"`
	Identity   *Identity
	Rules      []Rule    `gorm:"constraint:OnDelete:CASCADE"`
	DependsOn  []RuleSet `gorm:"many2many:RuleSetDependencies;constraint:OnDelete:CASCADE"`
}

func (r *RuleSet) Builtin() bool {
	return r.UUID != nil
}

//
// BeforeUpdate hook to avoid cyclic dependencies.
func (r *RuleSet) BeforeUpdate(db *gorm.DB) (err error) {
	seen := make(map[uint]bool)
	var nextDeps []RuleSet
	var nextRuleSetIDs []uint
	for _, dep := range r.DependsOn {
		nextRuleSetIDs = append(nextRuleSetIDs, dep.ID)
	}
	for len(nextRuleSetIDs) != 0 {
		result := db.Preload("DependsOn").Where("ID IN ?", nextRuleSetIDs).Find(&nextDeps)
		if result.Error != nil {
			err = result.Error
			return
		}
		nextRuleSetIDs = nextRuleSetIDs[:0]
		for _, nextDep := range nextDeps {
			for _, dep := range nextDep.DependsOn {
				if seen[dep.ID] {
					continue
				}
				if dep.ID == r.ID {
					err = DependencyCyclicError{}
					return
				}
				seen[dep.ID] = true
				nextRuleSetIDs = append(nextRuleSetIDs, dep.ID)
			}
		}
	}

	return
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
