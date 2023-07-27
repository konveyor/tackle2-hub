package filter

import (
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"reflect"
	"strconv"
	"strings"
)

type Any []interface{}

type And []interface{}

type Filter struct {
	predicates []string
}

func (f *Filter) String() (s string) {
	s = strings.Join(f.predicates, ",")
	return
}

func (f *Filter) Equals(field string, value interface{}) (self *Filter) {
	f.add(field, value, qf.EQ)
	return f
}

func (f *Filter) NotEquals(field string, value interface{}) (self *Filter) {
	f.add(field, value, qf.NOT)
	return f
}

func (f *Filter) Like(field string, value interface{}) *Filter {
	f.add(field, value, qf.LIKE)
	return f
}

func (f *Filter) GreaterThan(field string, value interface{}) (self *Filter) {
	f.add(field, value, qf.GT)
	return f
}

func (f *Filter) LessThan(field string, value interface{}) (self *Filter) {
	f.add(field, value, qf.GT)
	return f
}

func (f *Filter) add(field string, value interface{}, operator byte) (self *Filter) {
	self = f
	pv := f.valueOf(value)
	pv = field + string(operator) + pv
	f.predicates = append(
		f.predicates,
		pv)
	return
}

func (f *Filter) valueOf(object interface{}) (result string) {
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
			item := f.valueOf(value.Index(i).Interface())
			items = append(items, item)
		}
		var operator string
		switch object.(type) {
		case Any:
			operator = string(qf.OR)
		case And:
			operator = string(qf.AND)
		default:
			operator = string(qf.OR)
		}
		result = strings.Join(items, operator)
		result = "(" + result + ")"
	}
	return
}

// filter.Add(Field("name").Equals(dd)

func Field(name string) (p *Predicate) {
	p = &Predicate{field: name}
	return
}

type Predicate struct {
	field    string
	operator int32
	value    interface{}
}

func (p *Predicate) String() (s string) {
	pv := p.valueOf(p.value)
	s = p.field + string(p.operator) + pv
	return
}

func (p *Predicate) Equals(object interface{}) {
	p.operator = qf.EQ
	p.value = p.valueOf(object)
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
			operator = string(qf.OR)
		case And:
			operator = string(qf.AND)
		default:
			operator = string(qf.OR)
		}
		result = strings.Join(items, operator)
		result = "(" + result + ")"
	}
	return
}
