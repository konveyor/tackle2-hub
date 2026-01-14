package model

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/konveyor/tackle2-hub/internal/jsd"
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
	mp := jsd.JsonSafe(fieldValue)
	switch d := mp.(type) {
	case Data:
		mp = d.Any
	default:
	}
	v, err = json.Marshal(mp)
	return
}
