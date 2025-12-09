package api

// Dependency REST resource.
type Dependency struct {
	Resource `yaml:",inline"`
	To       Ref `json:"to"`
	From     Ref `json:"from"`
}
