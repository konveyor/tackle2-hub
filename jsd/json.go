package jsd

import (
	"fmt"
	"reflect"
	"time"
)

type Map = map[string]any

// JsonSafe ensures the object can be encoded as json.
// The YAML decoder can produce map[any]any which is not valid for json.
// Convert map[any]any to map[string]any as needed.
func JsonSafe(in any) (out any) {
	self := JsonSafe
	defer func() {
		r := recover()
		if r != nil {
			out = in
		}
	}()
	if in == nil {
		return
	}
	switch in.(type) {
	case time.Time, *time.Time:
		out = in
		return
	}
	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		out = reflect.New(t).Interface()
		nt := reflect.TypeOf(out)
		nt = nt.Elem()
		nv := reflect.ValueOf(out)
		nv = nv.Elem()
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			fv := v.Field(i)
			if !ft.IsExported() {
				continue
			}
			var object any
			switch fv.Kind() {
			case reflect.Ptr:
				if !v.IsNil() {
					object = fv.Elem().Interface()
				}
			default:
				object = fv.Interface()
			}
			object = self(object)
			ft = nt.Field(i)
			fv = nv.Field(i)
			x := reflect.ValueOf(object)
			fv.Set(x)
		}
		v = reflect.ValueOf(out)
		v = v.Elem()
		out = v.Interface()
	case reflect.Slice:
		list := make([]any, 0)
		for i := 0; i < v.Len(); i++ {
			v := v.Index(i)
			var object any
			switch v.Kind() {
			case reflect.Ptr:
				if !v.IsNil() {
					object = v.Elem().Interface()
				}
			default:
				object = v.Interface()
			}
			object = self(object)
			list = append(list, object)
		}
		out = list
	case reflect.Map:
		mp := make(map[string]any)
		for _, k := range v.MapKeys() {
			v := v.MapIndex(k)
			var object any
			switch v.Kind() {
			case reflect.Ptr:
				if !v.IsNil() {
					object = v.Elem().Interface()
				}
			default:
				object = v.Interface()
			}
			object = self(object)
			key := fmt.Sprintf("%v", k.Interface())
			mp[key] = object
		}
		out = mp
	default:
		out = in
	}

	return
}
