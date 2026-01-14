package api

// RuleSet REST resource.
type RuleSet struct {
	Resource    `yaml:",inline"`
	Kind        string      `json:"kind,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Rules       []Rule      `json:"rules"`
	Repository  *Repository `json:"repository,omitempty"`
	Identity    *Ref        `json:"identity,omitempty"`
	DependsOn   []Ref       `json:"dependsOn" yaml:"dependsOn"`
}

// Rule - REST Resource.
type Rule struct {
	Resource    `yaml:",inline"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	File        *Ref     `json:"file,omitempty"`
}

// Target REST resource.
type Target struct {
	Resource    `yaml:",inline"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Provider    string        `json:"provider,omitempty" yaml:",omitempty"`
	Choice      bool          `json:"choice,omitempty" yaml:",omitempty"`
	Custom      bool          `json:"custom,omitempty" yaml:",omitempty"`
	Labels      []TargetLabel `json:"labels"`
	Image       Ref           `json:"image"`
	RuleSet     *RuleSet      `json:"ruleset,omitempty" yaml:"ruleset,omitempty"`
}

// TargetLabel label.
type TargetLabel struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}
