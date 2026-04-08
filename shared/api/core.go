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

// APIKey REST resource.
type APIKey struct {
	Resource   `yaml:",inline"`
	Userid     string    `json:"userid,omitempty" yaml:"userid,omitempty"`
	Password   string    `json:"password,omitempty" yaml:"password,omitempty"`
	Digest     string    `json:"digest,omitempty" yaml:"digest,omitempty"`
	Secret     string    `json:"secret,omitempty" yaml:"secret,omitempty"`
	Lifespan   int       `json:"lifespan,omitempty" yaml:"lifespan,omitempty"`
	Expiration time.Time `json:"expiration,omitempty" yaml:"expiration,omitempty"`
	Expired    bool      `json:"expired,omitempty" yaml:"expired,omitempty"`
	User       *Ref      `json:"user,omitempty" yaml:"user,omitempty"`
	Task       *Ref      `json:"task,omitempty" yaml:"task,omitempty"`
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

// IdpIdentity REST resource.
type IdpIdentity struct {
	Resource          `yaml:",inline"`
	Provider          string    `json:"provider" binding:"required"`
	Subject           string    `json:"subject" binding:"required"`
	RefreshToken      string    `json:"refreshToken" binding:"required"`
	Expiration        time.Time `json:"expiration"`
	LastAuthenticated time.Time `json:"lastAuthenticated"`
	LastRefreshed     time.Time `json:"lastRefreshed"`
	User              *Ref      `json:"user" binding:"required"`
}

// User REST resource.
type User struct {
	Resource `yaml:",inline"`
	Subject  string `json:"subject"`
	Userid   string `json:"userid" binding:"required"`
	Password string `json:"password" binding:"required,max=72"`
	Email    string `json:"email" binding:"required"`
	Roles    []Ref  `json:"roles"`
}

// Role REST resource.
type Role struct {
	Resource    `yaml:",inline"`
	Name        string `json:"name" binding:"required"`
	Permissions []Ref  `json:"permissions"`
}

// Permission REST resource.
type Permission struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Scope    string `json:"scope" binding:"required"`
}
