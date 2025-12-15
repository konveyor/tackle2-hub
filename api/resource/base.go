package resource

import (
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

func ref(id uint, m any) (r Ref) {
	r.ID = id
	r.Name = reflect.NameOf(m)
	return
}

type Resource = api.Resource

// Ref type alias to shared API.
type Ref = api.Ref

// Map unstructured object.
type Map = api.Map

// TagRef type alias to shared API.
type TagRef = api.TagRef

// Repository REST nested resource.
type Repository = api.Repository

// IdentityRef type alias to shared API.
type IdentityRef = api.IdentityRef

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
