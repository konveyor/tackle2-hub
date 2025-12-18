package api

import (
	"strings"
	"time"
)

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

// TagRef represents a reference to a Tag.
// Contains the tag ID, name, tag source.
type TagRef struct {
	ID      uint   `json:"id" binding:"required"`
	Name    string `json:"name"`
	Source  string `json:"source,omitempty" yaml:"source,omitempty"`
	Virtual bool   `json:"virtual,omitempty" yaml:"virtual,omitempty"`
}

// IdentityRef application identity ref.
type IdentityRef struct {
	ID   uint   `json:"id" binding:"required"`
	Role string `json:"role" binding:"required"`
	Name string `json:"name"`
}

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

// Tag REST resource.
type Tag struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Category Ref    `json:"category" binding:"required"`
}

// Stakeholder REST resource.
type Stakeholder struct {
	Resource         `yaml:",inline"`
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Groups           []Ref  `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	BusinessServices []Ref  `json:"businessServices" yaml:"businessServices"`
	JobFunction      *Ref   `json:"jobFunction" yaml:"jobFunction"`
	Owns             []Ref  `json:"owns"`
	Contributes      []Ref  `json:"contributes"`
	MigrationWaves   []Ref  `json:"migrationWaves" yaml:"migrationWaves"`
}

// StakeholderGroup REST resource.
type StakeholderGroup struct {
	Resource       `yaml:",inline"`
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Stakeholders   []Ref  `json:"stakeholders"`
	MigrationWaves []Ref  `json:"migrationWaves" yaml:"migrationWaves"`
}

// JobFunction REST resource.
type JobFunction struct {
	Resource     `yaml:",inline"`
	Name         string `json:"name" binding:"required"`
	Stakeholders []Ref  `json:"stakeholders"`
}

// BusinessService REST resource.
type BusinessService struct {
	Resource    `yaml:",inline"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Stakeholder *Ref   `json:"owner" yaml:"owner"`
}

// Dependency REST resource.
type Dependency struct {
	Resource `yaml:",inline"`
	To       Ref `json:"to"`
	From     Ref `json:"from"`
}

// Review REST resource.
type Review struct {
	Resource            `yaml:",inline"`
	BusinessCriticality uint   `json:"businessCriticality" yaml:"businessCriticality"`
	EffortEstimate      string `json:"effortEstimate" yaml:"effortEstimate"`
	ProposedAction      string `json:"proposedAction" yaml:"proposedAction"`
	WorkPriority        uint   `json:"workPriority" yaml:"workPriority"`
	Comments            string `json:"comments"`
	Application         *Ref   `json:"application,omitempty" binding:"required_without=Archetype,excluded_with=Archetype"`
	Archetype           *Ref   `json:"archetype,omitempty" binding:"required_without=Application,excluded_with=Application"`
}

// CopyRequest REST resource.
type CopyRequest struct {
	SourceReview       uint   `json:"sourceReview" binding:"required"`
	TargetApplications []uint `json:"targetApplications" binding:"required"`
}

// MigrationWave REST Resource
type MigrationWave struct {
	Resource          `yaml:",inline"`
	Name              string    `json:"name"`
	StartDate         time.Time `json:"startDate" yaml:"startDate" binding:"required"`
	EndDate           time.Time `json:"endDate" yaml:"endDate" binding:"required,gtfield=StartDate"`
	Applications      []Ref     `json:"applications"`
	Stakeholders      []Ref     `json:"stakeholders"`
	StakeholderGroups []Ref     `json:"stakeholderGroups" yaml:"stakeholderGroups"`
}

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
