package model

//
// Analysis report.
type Analysis struct {
	Model
	Effort        int
	Issues        []Issue          `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies  []TechDependency `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID uint             `gorm:"index;not null"`
	Application   *Application
}

//
// TechDependency report dependency.
type TechDependency struct {
	Model
	Provider   string `gorm:"index:depA;not null;default:''"`
	Name       string `gorm:"index:depA;not null"`
	Version    string `gorm:"index:depA"`
	SHA        string `gorm:"index:depA"`
	Indirect   bool
	Labels     JSON `gorm:"type:json"`
	AnalysisID uint `gorm:"index;not null"`
	Analysis   *Analysis
}

//
// Issue report issue (violation).
type Issue struct {
	Model
	RuleSet     string `gorm:"index:issueA;not null"`
	Rule        string `gorm:"index:issueA;not null"`
	Name        string `gorm:"index"`
	Description string
	Category    string     `gorm:"index;not null"`
	Incidents   []Incident `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE"`
	Links       JSON       `gorm:"type:json"`
	Facts       JSON       `gorm:"type:json"`
	Labels      JSON       `gorm:"type:json"`
	Effort      int        `gorm:"index;not null"`
	AnalysisID  uint       `gorm:"index;not null"`
	Analysis    *Analysis
}

//
// RuleId returns unique rule ID.
func (m *Issue) RuleId() (id string) {
	return m.RuleSet + "." + m.Rule
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
