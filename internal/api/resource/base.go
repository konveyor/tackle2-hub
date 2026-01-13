package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/model/reflect"
	api2 "github.com/konveyor/tackle2-hub/shared/api"
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

func baseWith(r *api2.Resource, m *model.Model) {
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
type Resource = api2.Resource

// Ref REST resource.
type Ref = api2.Ref

// Map REST resource.
type Map = api2.Map

// TagRef REST resource.
type TagRef = api2.TagRef

// Repository REST resource.
type Repository = api2.Repository

// IdentityRef REST resource.
type IdentityRef = api2.IdentityRef

// InExList REST resource.
type InExList = model.InExList
