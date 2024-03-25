package v1alpha1

// Selector
// tag:category=tag
// platform:target=kind
type Selector struct {
	Match      string `json:"match,omitempty"`
	Name       string `json:"name,omitempty"`
	Capability string `json:"capability,omitempty"`
}
