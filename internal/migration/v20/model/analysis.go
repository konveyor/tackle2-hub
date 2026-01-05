package model

import (
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	"gorm.io/gorm"
)

// Analysis report.
type Analysis struct {
	Model
	Effort        int
	Commit        string
	Archived      bool
	Summary       []ArchivedInsight `gorm:"type:json;serializer:json"`
	Insights      []Insight         `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies  []TechDependency  `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID uint              `gorm:"index;not null"`
	Application   *Application
}

// TechDependency report dependency.
type TechDependency struct {
	Model
	Provider   string `gorm:"uniqueIndex:depA"`
	Name       string `gorm:"uniqueIndex:depA"`
	Version    string `gorm:"uniqueIndex:depA"`
	SHA        string `gorm:"uniqueIndex:depA"`
	Indirect   bool
	Labels     []string `gorm:"type:json;serializer:json"`
	AnalysisID uint     `gorm:"index;uniqueIndex:depA;not null"`
	Analysis   *Analysis
}

// Insight report insights.
type Insight struct {
	Model
	RuleSet     string `gorm:"uniqueIndex:insightA;not null"`
	Rule        string `gorm:"uniqueIndex:insightA;not null"`
	Name        string `gorm:"index"`
	Description string
	Category    string     `gorm:"index;not null"`
	Incidents   []Incident `gorm:"foreignKey:InsightID;constraint:OnDelete:CASCADE"`
	Links       []Link     `gorm:"type:json;serializer:json"`
	Facts       json.Map   `gorm:"type:json;serializer:json"`
	Labels      []string   `gorm:"type:json;serializer:json"`
	Effort      int        `gorm:"index;not null"`
	AnalysisID  uint       `gorm:"index;uniqueIndex:insightA;not null"`
	Analysis    *Analysis
}

// Incident report an issue incident.
type Incident struct {
	Model
	File      string `gorm:"index;not null"`
	Line      int
	Message   string
	CodeSnip  string
	Facts     json.Map `gorm:"type:json;serializer:json"`
	InsightID uint     `gorm:"index;not null"`
	Insight   *Insight
}

// RuleSet - Analysis ruleset.
type RuleSet struct {
	Model
	UUID        *string `gorm:"uniqueIndex"`
	Kind        string
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Repository  Repository `gorm:"type:json;serializer:json"`
	IdentityID  *uint      `gorm:"index"`
	Identity    *Identity
	Rules       []Rule    `gorm:"constraint:OnDelete:CASCADE"`
	DependsOn   []RuleSet `gorm:"many2many:RuleSetDependencies;constraint:OnDelete:CASCADE"`
}

func (r *RuleSet) Builtin() bool {
	return r.UUID != nil
}

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

// Rule - Analysis rule.
type Rule struct {
	Model
	Name        string
	Description string
	Labels      []string `gorm:"type:json;serializer:json"`
	RuleSetID   uint     `gorm:"uniqueIndex:RuleA;not null"`
	RuleSet     *RuleSet
	FileID      *uint `gorm:"uniqueIndex:RuleA" ref:"file"`
	File        *File
}

// Target - analysis rule selector.
type Target struct {
	Model
	UUID        *string `gorm:"uniqueIndex"`
	Name        string  `gorm:"uniqueIndex;not null"`
	Description string
	Provider    string
	Choice      bool
	Labels      []TargetLabel `gorm:"type:json;serializer:json"`
	ImageID     uint          `gorm:"index" ref:"file"`
	Image       *File
	RuleSetID   *uint `gorm:"index"`
	RuleSet     *RuleSet
}

func (r *Target) Builtin() bool {
	return r.UUID != nil
}

//
// JSON Fields.
//

// ArchivedInsight resource created when issues are archived.
type ArchivedInsight struct {
	RuleSet     string `json:"ruleSet"`
	Rule        string `json:"rule"`
	Name        string `json:"name,omitempty" yaml:",omitempty"`
	Description string `json:"description,omitempty" yaml:",omitempty"`
	Category    string `json:"category"`
	Effort      int    `json:"effort"`
	Incidents   int    `json:"incidents"`
}

// Link URL link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// TargetLabel - label format specific to Targets
type TargetLabel struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}
