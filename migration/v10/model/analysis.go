package model

//
// Analysis report.
type Analysis struct {
	Model
	Effort        int
	Archived      JSON             `gorm:"type:json"`
	Issues        []Issue          `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies  []TechDependency `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID uint             `gorm:"index;not null"`
	Application   *Application
}

//
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

//
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

//
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

//
// Link URL link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

//
// ArchivedAIssue resource created when issues are archived.
type ArchivedAIssue struct {
	RuleSet     string `json:"ruleSet"`
	Rule        string `json:"rule"`
	Name        string `json:"name,omitempty" yaml:",omitempty"`
	Description string `json:"description,omitempty" yaml:",omitempty"`
	Category    string `json:"category"`
	Effort      int    `json:"effort"`
	Incidents   int    `json:"incidents"`
}
