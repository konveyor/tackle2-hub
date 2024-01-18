package model

import "gorm.io/gorm"

// Analysis report.
type Analysis struct {
	Model
	Effort        int
	Archived      bool             `json:"archived"`
	Summary       JSON             `gorm:"type:json"`
	Issues        []Issue          `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies  []TechDependency `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID uint             `gorm:"index;not null"`
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
	Labels     JSON `gorm:"type:json"`
	AnalysisID uint `gorm:"index;uniqueIndex:depA;not null"`
	Analysis   *Analysis
}

// Issue report issue (violation).
type Issue struct {
	Model
	RuleSet     string `gorm:"uniqueIndex:issueA;not null"`
	Rule        string `gorm:"uniqueIndex:issueA;not null"`
	Name        string `gorm:"index"`
	Description string
	Category    string     `gorm:"index;not null"`
	Incidents   []Incident `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE"`
	Links       JSON       `gorm:"type:json"`
	Facts       JSON       `gorm:"type:json"`
	Labels      JSON       `gorm:"type:json"`
	Effort      int        `gorm:"index;not null"`
	AnalysisID  uint       `gorm:"index;uniqueIndex:issueA;not null"`
	Analysis    *Analysis
}

// Incident report an issue incident.
type Incident struct {
	Model
	File     string `gorm:"index;not null"`
	Line     int
	Message  string
	CodeSnip string
	Facts    JSON `gorm:"type:json"`
	IssueID  uint `gorm:"index;not null"`
	Issue    *Issue
}

// Link URL link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// ArchivedIssue resource created when issues are archived.
type ArchivedIssue struct {
	RuleSet     string `json:"ruleSet"`
	Rule        string `json:"rule"`
	Name        string `json:"name,omitempty" yaml:",omitempty"`
	Description string `json:"description,omitempty" yaml:",omitempty"`
	Category    string `json:"category"`
	Effort      int    `json:"effort"`
	Incidents   int    `json:"incidents"`
}

// RuleSet - Analysis ruleset.
type RuleSet struct {
	Model
	UUID        *string `gorm:"uniqueIndex"`
	Kind        string
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Repository  JSON  `gorm:"type:json"`
	IdentityID  *uint `gorm:"index"`
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
	Labels      JSON `gorm:"type:json"`
	RuleSetID   uint `gorm:"uniqueIndex:RuleA;not null"`
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
	Choice      bool
	Labels      JSON `gorm:"type:json"`
	ImageID     uint `gorm:"index" ref:"file"`
	Image       *File
	RuleSetID   *uint `gorm:"index"`
	RuleSet     *RuleSet
}

func (r *Target) Builtin() bool {
	return r.UUID != nil
}
