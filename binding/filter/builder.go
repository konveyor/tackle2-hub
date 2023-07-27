package filter

import (
	qf "github.com/konveyor/tackle2-hub/api/filter"
	//qf "github.com/konveyor/tackle2-hub/api/filter"
	"reflect"
	"strconv"
	"strings"
)

const (
	EQ   = string(qf.EQ)
	NOT  = string(qf.NOT)
	GT   = string(qf.GT)
	LT   = string(qf.LT)
	LIKE = string(qf.LIKE)
	AND  = string(qf.AND)
	OR   = string(qf.OR)
)

//
// Or match any.
type Or []interface{}

//
// And match all.
type And []interface{}

//
// Filter builder.
type Filter struct {
	predicates []*Predicate
}

//
// And adds a predicate.
// Example: filter.And("name").Equals("Elmer")
func (f *Filter) And(field string) (p *Predicate) {
	p = &Predicate{
		field:    field,
		operator: EQ,
	}
	f.predicates = append(f.predicates, p)
	return p
}

//
// String returns string representation.
func (f *Filter) String() (s string) {
	var preds []string
	for _, p := range f.predicates {
		preds = append(preds, p.String())
	}
	s = strings.Join(preds, string(qf.COMMA))
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
	s = p.field + p.operator + p.value
	return
}

//
// Eq returns a (=) predicate.
func (p *Predicate) Eq(object interface{}) *Predicate {
	p.operator = EQ
	p.value = p.valueOf(object)
	return p
}

//
// NotEq returns a (!=) predicate.
func (p *Predicate) NotEq(object interface{}) *Predicate {
	p.operator = NOT + EQ
	p.value = p.valueOf(object)
	return p
}

//
// Like returns a (~) predicate.
func (p *Predicate) Like(object interface{}) *Predicate {
	p.operator = LIKE
	p.value = p.valueOf(object)
	return p
}

//
// Gt returns a (>) predicate.
func (p *Predicate) Gt(object interface{}) *Predicate {
	p.operator = GT
	p.value = p.valueOf(object)
	return p
}

//
// GtEq returns a (>=) predicate.
func (p *Predicate) GtEq(object interface{}) *Predicate {
	p.operator = GT + EQ
	p.value = p.valueOf(object)
	return p
}

//
// Lt returns a (<) predicate.
func (p *Predicate) Lt(object interface{}) *Predicate {
	p.operator = LT
	p.value = p.valueOf(object)
	return p
}

//
// LtEq returns a (<) predicate.
func (p *Predicate) LtEq(object interface{}) *Predicate {
	p.operator = LT + EQ
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
		case Or:
			operator = OR
		case And:
			operator = AND
		default:
			operator = OR
		}
		result = strings.Join(items, operator)
		result = "(" + result + ")"
	}
	return
}
