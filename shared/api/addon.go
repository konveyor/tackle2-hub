package api

import "github.com/konveyor/tackle2-hub/shared/api/k8s"

// Addon REST resource.
type Addon struct {
	Name       string        `json:"name"`
	Container  k8s.Container `json:"container"`
	Extensions []Extension   `json:"extensions,omitempty"`
	Metadata   any           `json:"metadata,omitempty"`
}

// Extension REST resource.
type Extension struct {
	Name         string        `json:"name"`
	Addon        string        `json:"addon"`
	Capabilities []string      `json:"capabilities,omitempty"`
	Container    k8s.Container `json:"container"`
	Metadata     any           `json:"metadata,omitempty"`
}
