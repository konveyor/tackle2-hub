package model

//
// Analysis report.
type Analysis struct {
	Model
	Issues        []AnalysisIssue      `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies  []AnalysisDependency `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID uint                 `gorm:"index;not null"`
	Application   *Application
}

//
// AnalysisDependency report dependency.
type AnalysisDependency struct {
	Model
	Name       string `gorm:"index:depA;not null"`
	Version    string `gorm:"index:depA"`
	SHA        string `gorm:"index:depA"`
	Indirect   bool
	Labels     JSON `gorm:"type:json"`
	AnalysisID uint `gorm:"index;not null"`
	Analysis   *Analysis
}

//
// AnalysisIssue report issue (violation).
type AnalysisIssue struct {
	Model
	RuleSet     string `gorm:"index:AnalysisIssueA;not null"`
	Rule        string `gorm:"index:AnalysisIssueA;not null"`
	Name        string `gorm:"index"`
	Description string
	Category    string             `gorm:"index;not null"`
	Incidents   []AnalysisIncident `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE"`
	Links       JSON               `gorm:"type:json"`
	Facts       JSON               `gorm:"type:json"`
	Labels      JSON               `gorm:"type:json"`
	Effort      int                `gorm:"index;not null"`
	AnalysisID  uint               `gorm:"index;not null"`
	Analysis    *Analysis
}

//
// RuleId returns unique rule ID.
func (m *AnalysisIssue) RuleId() (id string) {
	return m.RuleSet + "." + m.Rule
}

//
// AnalysisIncident report incident.
type AnalysisIncident struct {
	Model
	URI     string
	Message string
	Facts   JSON `gorm:"type:json"`
	IssueID uint `gorm:"index;not null"`
	Issue   *AnalysisIssue
}

//
// AnalysisLink report link.
type AnalysisLink struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}
