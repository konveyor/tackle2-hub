package api

import (
	"time"
)

// Setting REST Resource
type Setting struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type RestAPI struct {
	Version string   `json:"version,omitempty" yaml:",omitempty"`
	Routes  []string `json:"routes"`
}

type LatestSchema struct {
	Name       string `json:"name"`
	Definition Map    `json:"definition"`
}

// File REST resource.
type File struct {
	Resource   `yaml:",inline"`
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Encoding   string     `yaml:"encoding,omitempty"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

// Bucket REST resource.
type Bucket struct {
	Resource   `yaml:",inline"`
	Path       string     `json:"path"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

// Addon REST resource.
type Addon struct {
	Name       string      `json:"name"`
	Container  Map         `json:"container"`
	Extensions []Extension `json:"extensions,omitempty"`
	Metadata   any         `json:"metadata,omitempty"`
}

// Extension REST resource.
type Extension struct {
	Name         string   `json:"name"`
	Addon        string   `json:"addon"`
	Capabilities []string `json:"capabilities,omitempty"`
	Container    Map      `json:"container"`
	Metadata     any      `json:"metadata,omitempty"`
}

// Identity REST resource.
type Identity struct {
	Resource    `yaml:",inline"`
	Kind        string `json:"kind" binding:"required"`
	Default     bool   `json:"default"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Key         string `json:"key"`
	Settings    string `json:"settings"`
}

// Proxy REST resource.
type Proxy struct {
	Resource `yaml:",inline"`
	Enabled  bool     `json:"enabled"`
	Kind     string   `json:"kind" binding:"oneof=http https"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Excluded []string `json:"excluded"`
	Identity *Ref     `json:"identity"`
}

// Repository REST nested resource.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

// Login REST resource.
type Login struct {
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token"`
	Refresh  string `json:"refresh"`
	Expiry   int    `json:"expiry"`
}
