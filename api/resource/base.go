package resource

import (
	"github.com/konveyor/tackle2-hub/api/jsd"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/model/reflect"
	"github.com/konveyor/tackle2-hub/shared/api"
)

func newRef(id uint, name string) (r model.Ref) {
	r = model.Ref{
		ID:   id,
		Name: name,
	}
	return
}

func refPtr(id *uint, m any) (ref *Ref) {
	if id == nil {
		return
	}
	ref = &Ref{}
	ref.ID = *id
	ref.Name = reflect.NameOf(m)
	return
}

func idPtr(ref *Ref) (id *uint) {
	if ref != nil {
		id = &ref.ID
	}
	return
}

func baseWith(r *api.Resource, m *model.Model) {
	r.ID = m.ID
	r.CreateUser = m.CreateUser
	r.UpdateUser = m.UpdateUser
	r.CreateTime = m.CreateTime
}

// Resource REST resource.
type Resource api.Resource

// With updates the resource with the model.
func (r *Resource) With(m *model.Model) {
	r.ID = m.ID
	r.CreateUser = m.CreateUser
	r.UpdateUser = m.UpdateUser
	r.CreateTime = m.CreateTime
}

// ref with id and named model.
func (r *Resource) ref(id uint, m any) (ref Ref) {
	ref.ID = id
	ref.Name = r.nameOf(m)
	return
}

// refPtr with id and named model.
func (r *Resource) refPtr(id *uint, m any) (ref *Ref) {
	if id == nil {
		return
	}
	ref = &Ref{}
	ref.ID = *id
	ref.Name = r.nameOf(m)
	return
}

// idPtr extracts ref ID.
func (r *Resource) idPtr(ref *Ref) (id *uint) {
	if ref != nil {
		id = &ref.ID
	}
	return
}

// nameOf model.
func (r *Resource) nameOf(m any) (name string) {
	name = reflect.NameOf(m)
	return
}

// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is optional and read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name,omitempty"`
}

// With id and named model.
func (r *Ref) With(id uint, name string) {
	r.ID = id
	r.Name = name
}

// Map unstructured object.
type Map = jsd.Map

// TagRef represents a reference to a Tag.
// Contains the tag ID, name, tag source.
type TagRef struct {
	ID      uint   `json:"id" binding:"required"`
	Name    string `json:"name"`
	Source  string `json:"source,omitempty" yaml:"source,omitempty"`
	Virtual bool   `json:"virtual,omitempty" yaml:"virtual,omitempty"`
}

// With id and named model.
func (r *TagRef) With(id uint, name string, source string, virtual bool) {
	r.ID = id
	r.Name = name
	r.Source = source
	r.Virtual = virtual
}

// Repository REST nested resource.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

// IdentityRef represents an identity reference with role.
type IdentityRef struct {
	ID   uint   `json:"id" binding:"required"`
	Role string `json:"role" binding:"required"`
	Name string `json:"name"`
}

// AppTag represents application tag mapping.
type AppTag struct {
	ApplicationID uint
	TagID         uint
	Source        string
	Tag           *model.Tag
}

func (r *AppTag) with(m *model.ApplicationTag) {
	r.ApplicationID = m.ApplicationID
	r.Source = m.Source
	r.Tag = &m.Tag
	r.TagID = m.TagID
}

// InExList include/exclude list.
type InExList = model.InExList

// Tag Sources
const (
	SourceAssessment = "assessment"
	SourceArchetype  = "archetype"
)
