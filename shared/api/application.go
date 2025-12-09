package api

import "strings"

// Application REST resource.
type Application struct {
	Resource        `yaml:",inline"`
	Name            string        `json:"name" binding:"required"`
	Description     string        `json:"description"`
	Bucket          *Ref          `json:"bucket"`
	Repository      *Repository   `json:"repository"`
	Assets          *Repository   `json:"assets"`
	Binary          string        `json:"binary"`
	Coordinates     *Document     `json:"coordinates"`
	Review          *Ref          `json:"review"`
	Comments        string        `json:"comments"`
	Identities      []IdentityRef `json:"identities"`
	Tags            []TagRef      `json:"tags"`
	BusinessService *Ref          `json:"businessService" yaml:"businessService"`
	Owner           *Ref          `json:"owner"`
	Contributors    []Ref         `json:"contributors"`
	MigrationWave   *Ref          `json:"migrationWave" yaml:"migrationWave"`
	Platform        *Ref          `json:"platform"`
	Archetypes      []Ref         `json:"archetypes"`
	Assessments     []Ref         `json:"assessments"`
	Manifests       []Ref         `json:"manifests"`
	Assessed        bool          `json:"assessed"`
	Risk            string        `json:"risk"`
	Confidence      int           `json:"confidence"`
	Effort          int           `json:"effort"`
}

// Fact REST nested resource.
type Fact struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

// FactKey is a fact source and fact name separated by a colon.
type FactKey string

// Qualify qualifies the name with the source.
func (r *FactKey) Qualify(source string) {
	*r = FactKey(
		strings.Join(
			[]string{source, r.Name()},
			":"))
}

// Source returns the source portion of a fact key.
func (r FactKey) Source() (source string) {
	s, _, found := strings.Cut(string(r), ":")
	if found {
		source = s
	}
	return
}

// Name returns the name portion of a fact key.
func (r FactKey) Name() (name string) {
	_, n, found := strings.Cut(string(r), ":")
	if found {
		name = n
	} else {
		name = string(r)
	}
	return
}

// Stakeholders REST subresource.
type Stakeholders struct {
	Owner        *Ref  `json:"owner"`
	Contributors []Ref `json:"contributors"`
}

// AppTag is a lightweight representation of ApplicationTag model.
type AppTag struct {
	ApplicationID uint
	TagID         uint
	Source        string
	Tag           any
}

// IdentityRef application identity ref.
type IdentityRef struct {
	ID   uint   `json:"id" binding:"required"`
	Role string `json:"role" binding:"required"`
	Name string `json:"name"`
}
