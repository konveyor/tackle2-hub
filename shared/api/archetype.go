package api

// TargetProfile REST resource.
type TargetProfile struct {
	Resource        `yaml:",inline"`
	Name            string `json:"name" binding:"required"`
	Generators      []Ref  `json:"generators" binding:"dive"`
	AnalysisProfile *Ref   `json:"analysisProfile,omitempty" yaml:"analysisProfile,omitempty"`
}

// Archetype REST resource.
type Archetype struct {
	Resource          `yaml:",inline"`
	Name              string          `json:"name" yaml:"name"`
	Description       string          `json:"description" yaml:"description"`
	Comments          string          `json:"comments" yaml:"comments"`
	Tags              []TagRef        `json:"tags" yaml:"tags" binding:"dive"`
	Criteria          []TagRef        `json:"criteria" yaml:"criteria" binding:"dive"`
	Stakeholders      []Ref           `json:"stakeholders" yaml:"stakeholders" binding:"dive"`
	StakeholderGroups []Ref           `json:"stakeholderGroups" yaml:"stakeholderGroups" binding:"dive"`
	Applications      []Ref           `json:"applications" yaml:"applications" binding:"dive"`
	Assessments       []Ref           `json:"assessments" yaml:"assessments" binding:"dive"`
	Assessed          bool            `json:"assessed"`
	Risk              string          `json:"risk"`
	Confidence        int             `json:"confidence"`
	Review            *Ref            `json:"review"`
	Profiles          []TargetProfile `json:"profiles" yaml:",omitempty" binding:"dive"`
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
	Profiles    []Ref       `json:"profiles" binding:"dive"`
}
