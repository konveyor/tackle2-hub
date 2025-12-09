package api

import (
	"time"
)

// Import REST resource.
type Import map[string]any

// ImportSummary REST resource.
type ImportSummary struct {
	Resource       `yaml:",inline"`
	Filename       string    `json:"filename"`
	ImportStatus   string    `json:"importStatus" yaml:"importStatus"`
	ImportTime     time.Time `json:"importTime" yaml:"importTime"`
	ValidCount     int       `json:"validCount" yaml:"validCount"`
	InvalidCount   int       `json:"invalidCount" yaml:"invalidCount"`
	CreateEntities bool      `json:"createEntities" yaml:"createEntities"`
}
