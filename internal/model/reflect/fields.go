package reflect

import (
	"fmt"
	"reflect"
	"time"

	liberr "github.com/jortel/go-utils/error"
)

// Fields returns a map of fields.
// Used for:
// - db.Updates()
// - sorting
func Fields(m any) (mp map[string]any) {
	var inspect func(r any)
	inspect = func(r any) {
		mt := reflect.TypeOf(r)
		mv := reflect.ValueOf(r)
		if mt.Kind() == reflect.Ptr {
			mt = mt.Elem()
			mv = mv.Elem()
		}
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			if !ft.IsExported() {
				continue
			}
			switch fv.Kind() {
			case reflect.Ptr:
				pt := ft.Type.Elem()
				switch pt.Kind() {
				case reflect.Struct,
					reflect.Slice,
					reflect.Array:
					continue
				default:
					mp[ft.Name] = fv.Interface()
				}
			case reflect.Struct:
				if ft.Anonymous {
					inspect(fv.Addr().Interface())
					continue
				}
				inst := fv.Interface()
				switch inst.(type) {
				case time.Time:
					mp[ft.Name] = inst
				}
			case reflect.Array:
				continue
			default:
				mp[ft.Name] = fv.Interface()
			}
		}
	}
	mp = map[string]any{}
	inspect(m)
	return
}

// NameOf returns the name of a model.
func NameOf(m any) (name string) {
	mt := reflect.TypeOf(m)
	mv := reflect.ValueOf(m)
	if mv.IsNil() {
		return
	}
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fv := mv.Field(i)
		switch ft.Name {
		case "Name":
			name = fv.String()
			return
		}
	}
	return
}

// HasFields returns the validated field names.
// Used for:
// - db.Omit()
// - db.Select()
func HasFields(m any, in ...string) (out []string, err error) {
	defer func() {
		p := recover()
		if p != nil {
			if pe, cast := p.(error); cast {
				err = pe
			} else {
				err = fmt.Errorf(
					"(paniced) failed: %#v",
					p)
			}
		}
	}()
	mp := make(map[string]any)
	var inspect func(r any)
	inspect = func(r any) {
		mt := reflect.TypeOf(r)
		mv := reflect.ValueOf(r)
		if mt.Kind() == reflect.Ptr {
			mt = mt.Elem()
			mv = mv.Elem()
		}
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			if !ft.IsExported() {
				continue
			}
			switch fv.Kind() {
			case reflect.Ptr:
				if ft.Anonymous {
					inspect(fv.Interface())
					continue
				}
				inst := fv.Interface()
				mp[ft.Name] = inst
			case reflect.Struct:
				if ft.Anonymous {
					inspect(fv.Addr().Interface())
					continue
				}
				inst := fv.Interface()
				mp[ft.Name] = inst
			default:
				mp[ft.Name] = fv.Interface()
			}
		}
	}
	inspect(m)
	for _, name := range in {
		_, found := mp[name]
		if !found {
			err = &FieldNotValid{
				Kind: NameOf(m),
				Name: name,
			}
			err = liberr.Wrap(err)
			return
		}
	}
	out = in
	return
}
