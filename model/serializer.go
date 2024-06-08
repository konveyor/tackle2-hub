package model

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm/schema"
)

func init() {
	schema.RegisterSerializer("json", jsonSerializer{})
}

type jsonSerializer struct {
}

// Scan implements serializer.
func (r jsonSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue any) (err error) {
	fieldValue := reflect.New(field.FieldType)
	if dbValue != nil {
		var b []byte
		switch v := dbValue.(type) {
		case string:
			b = []byte(v)
		case []byte:
			b = v
		default:
			return fmt.Errorf("json: failed to decode: %#v", dbValue)
		}
		if len(b) > 0 {
			ptr := fieldValue.Interface()
			switch d := ptr.(type) {
			case *Data:
				ptr = &d.Any
			default:
			}
			err = json.Unmarshal(b, ptr)
			if err != nil {
				return
			}
		}
	}
	v := fieldValue.Elem()
	field.ReflectValueOf(ctx, dst).Set(v)
	return
}

// Value implements serializer.
func (r jsonSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue any) (v any, err error) {
	mp := r.jMap(fieldValue)
	switch d := mp.(type) {
	case Data:
		mp = d.Any
	default:
	}
	v, err = json.Marshal(mp)
	return
}

// jMap returns a map[string]any.
// The YAML decoder can produce map[any]any which is not valid for json.
// Converts map[any]any to map[string]any as needed.
func (r jsonSerializer) jMap(in any) (out any) {
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
		mp := make(map[string]any)
		for i := 0; i < t.NumField(); i++ {
			t := t.Field(i)
			v := v.Field(i)
			if !t.IsExported() {
				continue
			}
			var object any
			switch v.Kind() {
			case reflect.Ptr:
				if !v.IsNil() {
					object = v.Elem().Interface()
				}
			default:
				object = v.Interface()
			}
			object = r.jMap(object)
			if t.Anonymous {
				if m, cast := object.(map[string]any); cast {
					for k, v := range m {
						mp[k] = v
					}
				}
			} else {
				mp[t.Name] = object
			}
		}
		out = mp
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
			object = r.jMap(object)
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
			object = r.jMap(object)
			key := fmt.Sprintf("%v", k.Interface())
			mp[key] = object
		}
		out = mp
	default:
		out = in
	}

	return
}
