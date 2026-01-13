package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/model/reflect"
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

// Resource REST resource.
type Resource = api.Resource

// Ref REST resource.
type Ref = api.Ref

// Map REST resource.
type Map = api.Map

// TagRef REST resource.
type TagRef = api.TagRef

// Repository REST resource.
type Repository = api.Repository

// IdentityRef REST resource.
type IdentityRef = api.IdentityRef

// InExList REST resource.
type InExList = model.InExList
