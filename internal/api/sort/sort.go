package sort

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/model/reflect"
	"gorm.io/gorm"
)

// Clause sort clause.
type Clause struct {
	direction string
	name      string
}

// Sort provides sorting.
type Sort struct {
	fields  map[string]any
	alias   map[string]string
	clauses []Clause
}

// Add adds virtual field and aliases.
func (r *Sort) Add(name string, aliases ...string) {
	r.init()
	for _, alias := range aliases {
		alias = strings.ToLower(alias)
		r.fields[alias] = 0
		r.alias[alias] = name
	}
}

// With context.
func (r *Sort) With(ctx *gin.Context, m any) (err error) {
	param := ctx.Query("sort")
	if param == "" {
		return
	}
	r.init()
	r.inspect(m)
	for _, s := range strings.Split(param, ",") {
		clause := Clause{}
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		mark := strings.Index(s, ":")
		if mark == -1 {
			_, found := r.fields[s]
			if !found {
				err = &SortError{s}
				return
			}
			clause.name = r.resolved(s)
			r.clauses = append(
				r.clauses,
				clause)
		} else {
			d := strings.ToLower(s[:mark])
			s := s[mark+1:]
			if len(d) != 0 {
				if d[0] == 'd' {
					clause.direction = "DESC"
				}
			}
			_, found := r.fields[s]
			if !found {
				err = &SortError{s}
				return
			}
			clause.name = r.resolved(s)
			r.clauses = append(
				r.clauses,
				clause)
		}
	}
	return
}

// Sorted returns sorted DB.
func (r *Sort) Sorted(in *gorm.DB) (out *gorm.DB) {
	r.init()
	out = in
	if len(r.clauses) == 0 {
		return
	}
	clauses := []string{}
	for _, clause := range r.clauses {
		clauses = append(clauses, clause.name+" COLLATE NOCASE "+clause.direction)
	}
	out = out.Order(strings.Join(clauses, ","))
	return
}

// init allocate maps.
func (r *Sort) init() {
	if r.fields == nil {
		r.fields = make(map[string]any)
	}
	if r.alias == nil {
		r.alias = make(map[string]string)
	}
}

// inspect object and return fields.
func (r *Sort) inspect(m any) {
	r.init()
	for key, v := range reflect.Fields(m) {
		key = strings.ToLower(key)
		r.fields[key] = v
	}
	return
}

// resolved returns field names with alias resolved.
func (r *Sort) resolved(in string) (out string) {
	r.init()
	out = in
	alias, found := r.alias[in]
	if found {
		out = alias
	}
	return
}
