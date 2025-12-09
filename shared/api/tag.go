package api

// Tag REST resource.
type Tag struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Category Ref    `json:"category" binding:"required"`
}
