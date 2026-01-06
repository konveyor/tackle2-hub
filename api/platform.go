package api

// Platform REST resource.
type Platform struct {
	Resource     `yaml:",inline"`
	Kind         string `json:"kind" binding:"required"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Identity     *Ref   `json:"identity,omitempty" yaml:",omitempty"`
	Applications []Ref  `json:"applications,omitempty" yaml:",omitempty"`
}

// Manifest REST resource.
type Manifest struct {
	Resource    `yaml:",inline"`
	Content     Map `json:"content"`
	Secret      Map `json:"secret,omitempty" yaml:"secret,omitempty"`
	Application Ref `json:"application"`
}
