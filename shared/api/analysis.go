package api

// Analysis REST resource.
type Analysis struct {
	Resource     `yaml:",inline"`
	Application  Ref               `json:"application"`
	Effort       int               `json:"effort"`
	Commit       string            `json:"commit,omitempty" yaml:",omitempty"`
	Archived     bool              `json:"archived,omitempty" yaml:",omitempty"`
	Insights     []Insight         `json:"insights,omitempty" yaml:",omitempty"`
	Dependencies []TechDependency  `json:"dependencies,omitempty" yaml:",omitempty"`
	Summary      []ArchivedInsight `json:"summary,omitempty" yaml:",omitempty" swaggertype:"object"`
}

// Insight REST resource.
type Insight struct {
	Resource    `yaml:",inline"`
	Analysis    uint       `json:"analysis"`
	RuleSet     string     `json:"ruleset" binding:"required"`
	Rule        string     `json:"rule" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty" yaml:",omitempty"`
	Category    string     `json:"category,omitempty" yaml:",omitempty"`
	Effort      int        `json:"effort,omitempty" yaml:",omitempty"`
	Incidents   []Incident `json:"incidents,omitempty" yaml:",omitempty"`
	Links       []Link     `json:"links,omitempty" yaml:",omitempty"`
	Facts       Map        `json:"facts,omitempty" yaml:",omitempty"`
	Labels      []string   `json:"labels"`
}

// TechDependency REST resource.
type TechDependency struct {
	Resource `yaml:",inline"`
	Analysis uint     `json:"analysis"`
	Provider string   `json:"provider" yaml:",omitempty"`
	Name     string   `json:"name" binding:"required"`
	Version  string   `json:"version,omitempty" yaml:",omitempty"`
	Indirect bool     `json:"indirect,omitempty" yaml:",omitempty"`
	Labels   []string `json:"labels,omitempty" yaml:",omitempty"`
	SHA      string   `json:"sha,omitempty" yaml:",omitempty"`
}

// Incident REST resource.
type Incident struct {
	Resource `yaml:",inline"`
	Insight  uint   `json:"insight"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	CodeSnip string `json:"codeSnip" yaml:"codeSnip"`
	Facts    Map    `json:"facts"`
}

// Link analysis report link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty" yaml:",omitempty"`
}

// ArchivedInsight created when insights are archived.
type ArchivedInsight struct {
	RuleSet     string `json:"ruleSet"`
	Rule        string `json:"rule"`
	Name        string `json:"name,omitempty" yaml:",omitempty"`
	Description string `json:"description,omitempty" yaml:",omitempty"`
	Category    string `json:"category"`
	Effort      int    `json:"effort"`
	Incidents   int    `json:"incidents"`
}

// RuleReport REST resource.
type RuleReport struct {
	RuleSet      string   `json:"ruleset"`
	Rule         string   `json:"rule"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Effort       int      `json:"effort"`
	Labels       []string `json:"labels"`
	Links        []Link   `json:"links"`
	Applications int      `json:"applications"`
}

// InsightReport REST resource.
type InsightReport struct {
	ID          uint     `json:"id"`
	RuleSet     string   `json:"ruleset"`
	Rule        string   `json:"rule"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Effort      int      `json:"effort"`
	Labels      []string `json:"labels"`
	Links       []Link   `json:"links"`
	Files       int      `json:"files"`
}

// InsightAppReport REST resource.
type InsightAppReport struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	BusinessService string `json:"businessService"`
	Effort          int    `json:"effort"`
	Incidents       int    `json:"incidents"`
	Files           int    `json:"files"`
	Insight         struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		RuleSet     string `json:"ruleset"`
		Rule        string `json:"rule"`
	} `json:"insight"`
}

// FileReport REST resource.
type FileReport struct {
	InsightID uint   `json:"insightId" yaml:"insightId"`
	File      string `json:"file"`
	Incidents int    `json:"incidents"`
	Effort    int    `json:"effort"`
}

// DepReport REST resource.
type DepReport struct {
	Provider     string   `json:"provider"`
	Name         string   `json:"name"`
	Labels       []string `json:"labels"`
	Applications int      `json:"applications"`
}

// DepAppReport REST resource.
type DepAppReport struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	BusinessService string `json:"businessService"`
	Dependency      struct {
		ID       uint     `json:"id"`
		Provider string   `json:"provider"`
		Name     string   `json:"name"`
		Version  string   `json:"version"`
		SHA      string   `json:"sha"`
		Indirect bool     `json:"indirect"`
		Labels   []string `json:"labels"`
	} `json:"dependency"`
}
