package api

// InExList include/exclude list.
type InExList struct {
	Included []string `json:"included"`
	Excluded []string `json:"excluded"`
}

// ApMode analysis mode.
type ApMode struct {
	WithDeps bool `json:"withDeps" yaml:"withDeps"`
}

// ApScope analysis scope.
type ApScope struct {
	WithKnownLibs bool     `json:"withKnownLibs" yaml:"withKnownLibs"`
	Packages      InExList `json:"packages,omitempty" yaml:",omitempty"`
}

// ApRules analysis rules.
type ApRules struct {
	Targets    []Ref       `json:"targets"`
	Labels     InExList    `json:"labels,omitempty" yaml:",omitempty"`
	Files      []Ref       `json:"files,omitempty" yaml:",omitempty"`
	Repository *Repository `json:"repository,omitempty" yaml:",omitempty"`
}

// AnalysisProfile REST resource.
type AnalysisProfile struct {
	Resource    `yaml:",inline"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty" yaml:",omitempty"`
	Mode        ApMode  `json:"mode"`
	Scope       ApScope `json:"scope"`
	Rules       ApRules `json:"rules"`
}
