package api

// JobFunction REST resource.
type JobFunction struct {
	Resource     `yaml:",inline"`
	Name         string `json:"name" binding:"required"`
	Stakeholders []Ref  `json:"stakeholders"`
}
