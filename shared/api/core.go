package api

import (
	"encoding/json"
	"time"

	"github.com/konveyor/tackle2-hub/shared/api/k8s"
)

// Resource base REST resource.
type Resource struct {
	ID         uint      `json:"id,omitempty" yaml:"id,omitempty"`
	CreateUser string    `json:"createUser" yaml:"createUser,omitempty"`
	UpdateUser string    `json:"updateUser" yaml:"updateUser,omitempty"`
	CreateTime time.Time `json:"createTime" yaml:"createTime,omitempty"`
}

// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is optional and read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name,omitempty"`
}

// Map unstructured object.
type Map map[string]any

// As convert the content into the object.
// The object must be a pointer.
func (m *Map) As(object any) (err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, object)
	return
}

// Setting REST Resource
type Setting struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// RestAPI resource.
type RestAPI struct {
	Version string   `json:"version,omitempty" yaml:",omitempty"`
	Routes  []string `json:"routes"`
}

// LatestSchema REST resource.
type LatestSchema struct {
	Name       string `json:"name"`
	Definition Map    `json:"definition"`
}

// Cache REST resource.
type Cache struct {
	Path     string `json:"path"`
	Capacity string `json:"capacity"`
	Used     string `json:"used"`
	Exists   bool   `json:"exists"`
}

// ConfigMap REST resource.
type ConfigMap struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

// Service REST resource.
type Service struct {
	Name  string `json:"name"`
	Route string `json:"route"`
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
