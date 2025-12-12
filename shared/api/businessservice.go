package api

// BusinessService REST resource.
type BusinessService struct {
	Resource    `yaml:",inline"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Stakeholder *Ref   `json:"owner" yaml:"owner"`
}
