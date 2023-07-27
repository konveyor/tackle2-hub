package filter

import (
	//qf "github.com/konveyor/tackle2-hub/api/filter"
	"reflect"
	"strconv"
	"strings"
)

//
// Any match any.
type Any []interface{}

//
// And match all.
type And []interface{}

//
// Filter builder.
type Filter struct {
	predicates []string
}

//
// String returns string representation.
func (f *Filter) String() (s string) {
	s = strings.Join(f.predicates, ",")
	return
}

//
// And adds a predicate.
func (f *Filter) And(p *Predicate) *Filter {
	f.Add(p)
	return f
}

//
// Add a predicate.
func (f *Filter) Add(p *Predicate) *Filter {
	f.predicates = append(
		f.predicates,
		p.String())
	return f
}

//
// Field returns a field predicate.
func Field(name string) (p *Predicate) {
	p = &Predicate{field: name}
	return
}

//
// Predicate is a filter query predicate.
type Predicate struct {
	field    string
	operator string
	value    string
}

//
// String returns a string representation of the predicate.
func (p *Predicate) String() (s string) {
	s = p.field + string(p.operator) + p.value
	return
}

//
// Equals returns a (=) predicate.
func (p *Predicate) Equals(object interface{}) *Predicate {
	p.operator = "="
	p.value = p.valueOf(object)
	return p
}

//
// NotEquals returns a (!=) predicate.
func (p *Predicate) NotEquals(object interface{}) *Predicate {
	p.operator = "!="
	p.value = p.valueOf(object)
	return p
}

//
// Like returns a (~) predicate.
func (p *Predicate) Like(object interface{}) *Predicate {
	p.operator = "~"
	p.value = p.valueOf(object)
	return p
}

//
// GreaterThan returns a (>) predicate.
func (p *Predicate) GreaterThan(object interface{}) *Predicate {
	p.operator = ">"
	p.value = p.valueOf(object)
	return p
}

//
// GreaterThanEq returns a (>=) predicate.
func (p *Predicate) GreaterThanEq(object interface{}) *Predicate {
	p.operator = ">="
	p.value = p.valueOf(object)
	return p
}

//
// LessThan returns a (<) predicate.
func (p *Predicate) LessThan(object interface{}) *Predicate {
	p.operator = "<"
	p.value = p.valueOf(object)
	return p
}

//
// LessThanEq returns a (<) predicate.
func (p *Predicate) LessThanEq(object interface{}) *Predicate {
	p.operator = "<="
	p.value = p.valueOf(object)
	return p
}

func (p *Predicate) valueOf(object interface{}) (result string) {
	kind := reflect.TypeOf(object).Kind()
	value := reflect.ValueOf(object)
	switch kind {
	case reflect.String:
		result = "'" + value.String() + "'"
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		n := value.Int()
		result = strconv.Itoa(int(n))
	case reflect.Bool:
		result = strconv.FormatBool(value.Bool())
	case reflect.Slice:
		var items []string
		for i := 0; i < value.Len(); i++ {
			item := p.valueOf(value.Index(i).Interface())
			items = append(items, item)
		}
		var operator string
		switch object.(type) {
		case Any:
			operator = "|"
		case And:
			operator = ","
		default:
			operator = "|"
		}
		result = strings.Join(items, operator)
		result = "(" + result + ")"
	}
	return
}
