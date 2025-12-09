package api

// Stakeholder REST resource.
type Stakeholder struct {
	Resource         `yaml:",inline"`
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Groups           []Ref  `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	BusinessServices []Ref  `json:"businessServices" yaml:"businessServices"`
	JobFunction      *Ref   `json:"jobFunction" yaml:"jobFunction"`
	Owns             []Ref  `json:"owns"`
	Contributes      []Ref  `json:"contributes"`
	MigrationWaves   []Ref  `json:"migrationWaves" yaml:"migrationWaves"`
}

// StakeholderGroup REST resource.
type StakeholderGroup struct {
	Resource       `yaml:",inline"`
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Stakeholders   []Ref  `json:"stakeholders"`
	MigrationWaves []Ref  `json:"migrationWaves" yaml:"migrationWaves"`
}
