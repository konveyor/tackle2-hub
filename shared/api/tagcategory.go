package api

// TagCategory REST resource.
type TagCategory struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Color    string `json:"colour" yaml:"colour"`
	Tags     []Ref  `json:"tags"`
	// Deprecated
	Username string `json:"username,omitempty"` // Deprecated
	Rank     uint   `json:"rank,omitempty"`     // Deprecated
}
