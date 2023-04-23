package model

import "strings"

//
// Analysis report.
type Analysis struct {
	Model
	RuleSets      []AnalysisRuleSet    `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies  []AnalysisDependency `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID uint
	Application   *Application
}

//
// AnalysisRuleSet report ruleset.
type AnalysisRuleSet struct {
	Model
	Name         string
	Description  string
	Technologies []AnalysisTechnology `gorm:"foreignKey:RuleSetID;constraint:OnDelete:CASCADE"`
	Issues       []AnalysisIssue      `gorm:"foreignKey:RuleSetID;constraint:OnDelete:CASCADE"`
	AnalysisID   uint
	Analysis     *Analysis
}

//
// AnalysisTechnology report technology ref.
type AnalysisTechnology struct {
	Model
	Name      string `gorm:"index:techA;not null"`
	Version   string `gorm:"index:techA;not null"`
	Source    bool
	RuleSetID uint
	RuleSet   *AnalysisRuleSet
}

//
// AnalysisDependency report dependency.
type AnalysisDependency struct {
	Model
	Name       string `gorm:"index:depA;not null"`
	Version    string `gorm:"index:depA"`
	Type       string `gorm:"index:depA"`
	SHA        string `gorm:"index:depA"`
	Indirect   bool
	AnalysisID uint
	Analysis   *Analysis
}

//
// Key used for comparison.
func (m *AnalysisDependency) Key() (s string) {
	s = strings.Join(
		[]string{
			m.Name,
			m.Version,
			m.Type,
			m.SHA,
		},
		":")
	return
}

//
// AnalysisIssue report issue (violation).
type AnalysisIssue struct {
	Model
	RuleID      string `gorm:"index;not null"`
	Description string
	Category    string             `gorm:"index;not null"`
	Incidents   []AnalysisIncident `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE"`
	Links       JSON               `gorm:"type:json"`
	Facts       JSON               `gorm:"type:json"`
	Effort      int                `gorm:"index;not null"`
	RuleSetID   uint
	RuleSet     *AnalysisRuleSet
}

//
// AnalysisIncident report incident.
type AnalysisIncident struct {
	Model
	URI     string
	Message string
	Facts   JSON `gorm:"type:json"`
	IssueID uint
	Issue   *AnalysisIssue
}

//
// AnalysisLink report link.
type AnalysisLink struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}
