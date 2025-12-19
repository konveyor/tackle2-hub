package api

// TargetProfile REST resource.
type TargetProfile struct {
	Resource        `yaml:",inline"`
	Name            string `json:"name" binding:"required"`
	Generators      []Ref  `json:"generators"`
	AnalysisProfile *Ref   `json:"analysisProfile,omitempty" yaml:"analysisProfile,omitempty"`
}

// Archetype REST resource.
type Archetype struct {
	Resource          `yaml:",inline"`
	Name              string          `json:"name" yaml:"name"`
	Description       string          `json:"description" yaml:"description"`
	Comments          string          `json:"comments" yaml:"comments"`
	Tags              []TagRef        `json:"tags" yaml:"tags"`
	Criteria          []TagRef        `json:"criteria" yaml:"criteria"`
	Stakeholders      []Ref           `json:"stakeholders" yaml:"stakeholders"`
	StakeholderGroups []Ref           `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	Applications      []Ref           `json:"applications" yaml:"applications"`
	Assessments       []Ref           `json:"assessments" yaml:"assessments"`
	Assessed          bool            `json:"assessed"`
	Risk              string          `json:"risk"`
	Confidence        int             `json:"confidence"`
	Review            *Ref            `json:"review"`
	Profiles          []TargetProfile `json:"profiles" yaml:",omitempty"`
}

// Generator REST resource.
type Generator struct {
	Resource    `yaml:",inline"`
	Kind        string      `json:"kind" binding:"required"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty" yaml:",omitempty"`
	Repository  *Repository `json:"repository"`
	Params      Map         `json:"params"`
	Values      Map         `json:"values"`
	Identity    *Ref        `json:"identity,omitempty" yaml:",omitempty"`
	Profiles    []Ref       `json:"profiles"`
}
