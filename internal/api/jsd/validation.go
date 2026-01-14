package jsd

import (
	"reflect"

	"github.com/konveyor/tackle2-hub/internal/jsd"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Settings = &settings.Settings

// New validator.
func New(client client.Client) (v *Validator) {
	v = &Validator{
		manager: jsd.New(client),
	}
	return
}

// Validator jsd validator.
type Validator struct {
	manager *jsd.Manager
}

// Validate validates the specified document based on schema.
func (v *Validator) Validate(r any) (err error) {
	if Settings.Hub.Disconnected {
		return
	}
	fields := v.fields(r)
	for _, f := range fields {
		if f != nil {
			err = f.Validate(v.manager)
			if err != nil {
				return
			}
		}
	}
	return
}

// fields returns resource `Document` fields.
func (v *Validator) fields(r any) (fields []*Document) {
	rt := reflect.TypeOf(r)
	rv := reflect.ValueOf(r)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)
		if !ft.IsExported() {
			continue
		}
		object := fv.Interface()
		switch d := object.(type) {
		case *Document:
			fields = append(fields, d)
		case Document:
			fields = append(fields, &d)
		}
	}
	return
}
