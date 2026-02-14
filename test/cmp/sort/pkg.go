package sort

import (
	"reflect"
	"sort"
)

var (
	Registered Map
)

func init() {
	Registered = make(Map)
}

// Add a sort mapping.
func Add(s Sort, values ...any) {
	Registered.Add(s, values...)
}

// Reset mapping.
func Reset() {
	Registered.Reset()
}

// Sort (slice) sort.
type Sort func(reflect.Value)

// Map of sort.
type Map map[reflect.Type]Sort

// Add a sort mapping.
func (m Map) Add(s Sort, values ...any) {
	for _, v := range values {
		m[reflect.TypeOf(v)] = s
	}
}

// Reset mapping.
func (m Map) Reset() {
	for k := range m {
		delete(m, k)
	}
}

// ById sorts list of structs with an ID field.
func ById(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.IsNil() {
		return
	}
	if v.Kind() != reflect.Slice {
		return
	}
	if v.Len() < 2 {
		return
	}
	first := v.Index(0)
	if first.Kind() != reflect.Struct {
		return
	}
	idField := first.FieldByName("ID")
	if !idField.IsValid() || !idField.CanUint() {
		return
	}
	sort.Slice(
		v.Interface(),
		func(i, j int) bool {
			a := v.Index(i).FieldByName("ID")
			b := v.Index(j).FieldByName("ID")
			return a.Uint() < b.Uint()
		})
	return
}

// ByName sorts list of structs with a Name field.
func ByName(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.IsNil() {
		return
	}
	if v.Kind() != reflect.Slice {
		return
	}
	if v.Len() < 2 {
		return
	}
	first := v.Index(0)
	if first.Kind() != reflect.Struct {
		return
	}
	nameField := first.FieldByName("Name")
	if !nameField.IsValid() {
		return
	}
	sort.Slice(
		v.Interface(),
		func(i, j int) bool {
			a := v.Index(i).FieldByName("Name")
			b := v.Index(j).FieldByName("Name")
			return a.String() < b.String()
		})
	return
}
