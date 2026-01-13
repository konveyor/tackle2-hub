package filter

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	api "github.com/konveyor/tackle2-hub/shared/api/filter"
	"gorm.io/gorm"
)

const (
	QueryParam = api.Filter
)

// New filter.
func New(ctx *gin.Context, assertions []Assert) (f Filter, err error) {
	p := Parser{}
	q := strings.Join(
		ctx.QueryArray(QueryParam),
		string(COMMA))
	f, err = p.Filter(q)
	if err != nil {
		return
	}
	err = f.Validate(assertions)
	return
}

// Filter is a collection of predicates.
type Filter struct {
	predicates []Predicate
}

// Validate -
func (f *Filter) Validate(assertions []Assert) (err error) {
	if len(assertions) == 0 {
		return
	}
	find := func(name string) (assert *Assert, found bool) {
		name = strings.ToLower(name)
		for i := range assertions {
			assert = &assertions[i]
			if strings.ToLower(assert.Field) == name {
				found = true
				break
			}
		}
		return
	}
	for _, p := range f.predicates {
		name := p.Field.Value
		v, found := find(name)
		if !found {
			err = Errorf("'%s' not supported.", name)
			return
		}
		err = v.assert(&p)
		if err != nil {
			break
		}
	}

	return
}

// Field returns a field.
func (f *Filter) Field(name string) (field Field, found bool) {
	fields := f.Fields(name)
	if len(fields) > 0 {
		field = fields[0]
		found = true
	}
	return
}

// Fields returns fields.
func (f *Filter) Fields(name string) (fields []Field) {
	name = strings.ToLower(name)
	for _, p := range f.predicates {
		if strings.ToLower(p.Field.Value) == name {
			f := Field{p}
			fields = append(fields, f)
		}
	}
	return
}

// Resource returns a filter scoped to resource.
func (f *Filter) Resource(r string) (filter Filter) {
	r = strings.ToLower(r)
	var predicates []Predicate
	for _, p := range f.predicates {
		field := Field{p}
		fr := field.Resource()
		fr = strings.ToLower(fr)
		if fr == r {
			p.Field.Value = field.Name()
			predicates = append(predicates, p)
		}
	}
	filter.predicates = predicates
	return
}

// Where applies (root) fields to the where clause.
func (f *Filter) Where(in *gorm.DB, selector ...string) (out *gorm.DB) {
	out = in
	fs := FieldSelector(selector)
	for _, p := range f.predicates {
		field := Field{p}
		if fs.Match(&field) {
			out = field.Where(out)
		}
	}
	return
}

// With return filter with selected predicates.
func (f *Filter) With(selector ...string) (out Filter) {
	fs := FieldSelector(selector)
	for _, p := range f.predicates {
		field := Field{p}
		if fs.Match(&field) {
			out.predicates = append(out.predicates, p)
		}
	}
	return
}

// Renamed return a filter with predicate field renamed.
func (f *Filter) Renamed(name, renamed string) (out Filter) {
	var predicates []Predicate
	for _, p := range f.predicates {
		if p.Field.Value == name {
			p.Field.Value = renamed
		}
		predicates = append(
			predicates,
			p)
	}
	out.predicates = predicates
	return
}

// Revalued return a filter with the named field value replaced.
func (f *Filter) Revalued(name string, value Value) (out Filter) {
	var predicates []Predicate
	for _, p := range f.predicates {
		if p.Field.Value == name {
			p.Value = value
		}
		predicates = append(
			predicates,
			p)
	}
	out.predicates = predicates
	return
}

// Delete specified fields.
func (f *Filter) Delete(name string) (found bool) {
	var wanted []Predicate
	for _, p := range f.predicates {
		if strings.ToLower(p.Field.Value) != name {
			wanted = append(wanted, p)
		} else {
			found = true
		}
	}
	f.predicates = wanted
	return
}

// Empty returns true when the filter has no predicates.
func (f *Filter) Empty() bool {
	return len(f.predicates) == 0
}

// FieldSelector fields.
// fields with '+' prefix are included.
// fields with '-' prefix are excluded.
// Fields scoped to a resource are excluded.
// An empty selector includes ALL.
type FieldSelector []string

// Match fields by qualified name.
func (r FieldSelector) Match(f *Field) (m bool) {
	if f.Resource() != "" {
		return
	}
	if len(r) == 0 {
		m = true
		return
	}
	included := make(map[string]byte)
	excluded := make(map[string]byte)
	for _, s := range r {
		s = strings.ToLower(s)
		if s != "" {
			switch s[0] {
			case '-':
				excluded[s[1:]] = 0
			case '+':
				included[s[1:]] = 0
			default:
				included[s] = 0
			}
		}
	}
	name := strings.ToLower(f.Field.Value)
	if _, found := excluded[name]; found {
		return
	}
	if len(included) == 0 {
		m = true
		return
	}
	_, m = included[name]
	return
}

// Field predicate.
type Field struct {
	Predicate
}

// Name returns the field name.
func (f *Field) Name() (s string) {
	_, s = f.split()
	return
}

// As returns the renamed field.
func (f *Field) As(s string) (named Field) {
	named = Field{f.Predicate}
	named.Field.Value = s
	return
}

// Resource returns the field resource.
func (f *Field) Resource() (s string) {
	s, _ = f.split()
	return
}

// Where updates the where clause.
func (f *Field) Where(in *gorm.DB) (out *gorm.DB) {
	sql, values := f.SQL()
	out = in.Where(sql, values...)
	return
}

// SQL builds SQL.
// Returns statement and values (for ?).
func (f *Field) SQL() (s string, vList []any) {
	name := f.Name()
	switch len(f.Value) {
	case 0:
	case 1:
		switch f.Operator.Value {
		case string(LIKE):
			v := strings.Replace(f.Value[0].Value, "*", "%", -1)
			vList = append(vList, v)
			s = strings.Join(
				[]string{
					name,
					f.operator(),
					"?",
				},
				" ")
		default:
			vList = append(vList, AsValue(f.Value[0]))
			s = strings.Join(
				[]string{
					name,
					f.operator(),
					"?",
				},
				" ")
		}
	default:
		if f.Value.Operator(AND) {
			// not supported.
			break
		}
		switch f.Operator.Value {
		case string(LIKE):
			s = "("
			var clauses []string
			for _, fx := range f.Expand() {
				sql, values := fx.SQL()
				vList = append(vList, values[0])
				clauses = append(clauses, sql)
			}
			s += strings.Join(clauses, " OR ")
			s += ")"
		default:
			values := f.Value.ByKind(LITERAL, STRING)
			var collection []any
			for i := range values {
				v := AsValue(values[i])
				collection = append(collection, v)
			}
			vList = append(vList, collection)
			s = strings.Join(
				[]string{
					name,
					f.operator(),
					"?",
				},
				" ")
		}
	}
	return
}

// Expand flattens a multi-value field and returns a Field for each value.
func (f *Field) Expand() (expanded []Field) {
	for _, v := range f.Value.ByKind(LITERAL, STRING) {
		expanded = append(
			expanded,
			Field{Predicate{
				Unused:   f.Predicate.Unused,
				Field:    f.Predicate.Field,
				Operator: f.Predicate.Operator,
				Value:    Value{v},
			}})
	}
	return
}

// split field name.
// format: resource.name
// The resource may be "" (anonymous).
// The (.) separator is escaped when preceded by (\).
func (f *Field) split() (relation string, name string) {
	s := f.Field.Value
	mark := strings.Index(s, ".")
	if mark == -1 {
		name = s
		return
	}
	if mark > 0 && s[mark-1] == '\\' {
		name = s[:mark-1] + s[mark:]
		return
	}
	relation = s[:mark]
	name = s[mark+1:]
	return
}

// operator returns SQL operator.
func (f *Field) operator() (s string) {
	switch len(f.Value) {
	case 1:
		s = f.Operator.Value
		switch s {
		case string(COLON):
			s = "="
		case string(LIKE):
			s = "LIKE"
		}
	default:
		s = "IN"
	}

	return
}

// Assert -
type Assert struct {
	Field string
	Kind  byte
	And   bool
}

// assert validation.
func (r *Assert) assert(p *Predicate) (err error) {
	name := p.Field.Value
	switch r.Kind {
	case LITERAL:
		switch p.Operator.Value {
		case string(LIKE):
			err = Errorf("'~' cannot be used with '%s'", name)
			return
		}
	}
	if !r.And {
		if (&Field{*p}).Value.Operator(AND) {
			err = Errorf("(,,) cannot be used with '%s'.", name)
			return
		}
	}
	return
}

// AsValue returns the real value.
func AsValue(t Token) (object any) {
	v := t.Value
	object = v
	switch t.Kind {
	case LITERAL:
		n, err := strconv.Atoi(v)
		if err == nil {
			object = n
			break
		}
		b, err := strconv.ParseBool(v)
		if err == nil {
			object = b
			break
		}
	default:
	}
	return
}
