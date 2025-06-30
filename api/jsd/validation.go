package jsd

import (
	reflect2 "reflect"

	"github.com/konveyor/tackle2-hub/jsd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	rt := reflect2.TypeOf(r)
	rv := reflect2.ValueOf(r)
	if rt.Kind() == reflect2.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)
		if !ft.IsExported() {
			continue
		}
		v := fv.Interface()
		switch d := v.(type) {
		case *Document:
			fields = append(fields, d)
		case Document:
			fields = append(fields, &d)
		}
	}
	return
}
