package sort

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

//
// Clause sort clause.
type Clause struct {
	direction string
	name      string
}

//
// Sort provides sorting.
type Sort struct {
	fields  map[string]bool
	clauses []Clause
}

//
// With context.
func (r *Sort) With(ctx *gin.Context, m interface{}) (err error) {
	param := ctx.Query("sort")
	if param == "" {
		return
	}
	r.fields = r.inspect(m)
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
			clause.name = s
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
			clause.name = s
			r.clauses = append(
				r.clauses,
				clause)
		}
	}
	return
}

//
// Sorted returns sorted DB.
func (r *Sort) Sorted(in *gorm.DB) (out *gorm.DB) {
	out = in
	if len(r.clauses) == 0 {
		return
	}
	clauses := []string{}
	for _, clause := range r.clauses {
		clauses = append(clauses, clause.name+" "+clause.direction)
	}
	out = out.Order(strings.Join(clauses, ","))
	return
}

//
// inspect object and return fields.
func (r *Sort) inspect(m interface{}) (fields map[string]bool) {
	fields = make(map[string]bool)
	var inspect func(r interface{})
	inspect = func(r interface{}) {
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
			case reflect.Struct:
				if ft.Anonymous {
					inspect(fv.Interface())
					continue
				}
				inst := fv.Interface()
				switch inst.(type) {
				case time.Time:
					fields[strings.ToLower(ft.Name)] = true
				}
			case reflect.Array,
				reflect.Slice,
				reflect.Ptr:
				continue
			default:
				fields[strings.ToLower(ft.Name)] = true
			}
		}
	}
	inspect(m)
	return
}
