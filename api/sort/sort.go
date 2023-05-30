package sort

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

//
// Names column map.
type Names map[string]string

//
// Clause sort clause.
type Clause struct {
	direction string
	name      string
}

//
// Sort provides sorting.
type Sort struct {
	names   Names
	clauses []Clause
}

//
// With context.
// Fields formats:
//   name
//   name|column
func (r *Sort) With(ctx *gin.Context, fields ...interface{}) (err error) {
	param := ctx.Query("sort")
	if param == "" {
		return
	}
	r.names = make(Names)
	for _, object := range fields {
		switch object.(type) {
		case string:
			s := object.(string)
			part := strings.SplitN(s, "|", 2)
			name := strings.ToLower(part[0])
			column := name
			if len(part) == 2 {
				column = part[1]
			}
			r.names[name] = column
		default:
			for _, s := range r.inspect(object) {
				s = strings.ToLower(s)
				r.names[s] = s
			}
		}
	}
	for _, s := range strings.Split(param, ",") {
		clause := Clause{}
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		mark := strings.Index(s, ":")
		if mark == -1 {
			column, found := r.names[s]
			if !found {
				err = &SortError{s}
				return
			}
			clause.name = column
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
			column, found := r.names[s]
			if !found {
				err = &SortError{s}
				return
			}
			clause.name = column
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
func (r *Sort) inspect(m interface{}) (fields []string) {
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
					fields = append(fields, ft.Name)
				}
			case reflect.Array,
				reflect.Slice,
				reflect.Ptr:
				continue
			default:
				fields = append(fields, ft.Name)
			}
		}
	}
	inspect(m)
	return
}
