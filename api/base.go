package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"io"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

var Log = logging.WithName("api")

//
// BaseHandler base handler.
type BaseHandler struct {
}

//
// DB return db client associated with the context.
func (h *BaseHandler) DB(ctx *gin.Context) (db *gorm.DB) {
	object, found := ctx.Get(CtxDB)
	if !found {
		panic(liberr.New("context: DB not found."))
	}
	db, cast := object.(*gorm.DB)
	if !cast {
		panic(liberr.New("context: DB not valid."))
	}
	return
}

//
// Client return the k8s client associated with the context.
func (h *BaseHandler) Client(ctx *gin.Context) (c client.Client) {
	object, found := ctx.Get(CtxClient)
	if !found {
		panic(liberr.New("context: k8s client not found."))
	}
	c, cast := object.(client.Client)
	if !cast {
		panic(liberr.New("context: k8s client not valid."))
	}
	return
}

//
// preLoad update DB to pre-load fields.
func (h *BaseHandler) preLoad(db *gorm.DB, fields ...string) (tx *gorm.DB) {
	tx = db
	for _, f := range fields {
		tx = tx.Preload(f)
	}

	return
}

//
// fields builds a map of fields.
func (h *BaseHandler) fields(m interface{}) (mp map[string]interface{}) {
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
			case reflect.Ptr:
				pt := ft.Type.Elem()
				switch pt.Kind() {
				case reflect.Struct, reflect.Slice, reflect.Array:
					continue
				default:
					mp[ft.Name] = fv.Interface()
				}
			case reflect.Struct:
				if ft.Anonymous {
					inspect(fv.Addr().Interface())
				}
			case reflect.Array:
				continue
			case reflect.Slice:
				inst := fv.Interface()
				switch inst.(type) {
				case []byte:
					mp[ft.Name] = fv.Interface()
				}
			default:
				mp[ft.Name] = fv.Interface()
			}
		}
	}
	mp = map[string]interface{}{}
	inspect(m)
	return
}

//
// pk returns the PK (ID) parameter.
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
	s := ctx.Param(ID)
	n, _ := strconv.Atoi(s)
	id = uint(n)
	return
}

//
// modBody updates the body using the `mod` function.
//   1. read the body.
//   2. mod()
//   3. write body.
func (h *BaseHandler) modBody(
	ctx *gin.Context,
	r interface{},
	mod func(bool) error) (err error) {
	//
	withBody := false
	if ctx.Request.ContentLength > 0 {
		withBody = true
		err = ctx.BindJSON(r)
		if err != nil {
			return
		}
	}
	err = mod(withBody)
	if err != nil {
		return
	}
	b, _ := json.Marshal(r)
	bfr := bytes.NewBuffer(b)
	ctx.Request.Body = io.NopCloser(bfr)
	return
}

//
// CurrentUser gets username from Keycloak auth token.
func (h *BaseHandler) CurrentUser(ctx *gin.Context) (user string) {
	user = ctx.GetString(auth.TokenUser)
	if user == "" {
		Log.Info("Failed to get current user.")
	}

	return
}

//
// HasScope determines if the token has the specified scope.
func (h *BaseHandler) HasScope(ctx *gin.Context, scope string) (b bool) {
	in := auth.BaseScope{}
	in.With(scope)
	if object, found := ctx.Get(auth.TokenScopes); found {
		if scopes, cast := object.([]auth.Scope); cast {
			for _, s := range scopes {
				b = s.Match(in.Resource, in.Method)
				if b {
					return
				}
			}
		}
	}
	return
}

//
// REST resource.
type Resource struct {
	ID         uint      `json:"id"`
	CreateUser string    `json:"createUser"`
	UpdateUser string    `json:"updateUser"`
	CreateTime time.Time `json:"createTime"`
}

//
// With updates the resource with the model.
func (r *Resource) With(m *model.Model) {
	r.ID = m.ID
	r.CreateUser = m.CreateUser
	r.UpdateUser = m.UpdateUser
	r.CreateTime = m.CreateTime
}

//
// ref with id and named model.
func (r *Resource) ref(id uint, m interface{}) (ref Ref) {
	ref.ID = id
	ref.Name = r.nameOf(m)
	return
}

//
// refPtr with id and named model.
func (r *Resource) refPtr(id *uint, m interface{}) (ref *Ref) {
	if id == nil {
		return
	}
	ref = &Ref{}
	ref.ID = *id
	ref.Name = r.nameOf(m)
	return
}

//
// idPtr extracts ref ID.
func (r *Resource) idPtr(ref *Ref) (id *uint) {
	if ref != nil {
		id = &ref.ID
	}
	return
}

//
// nameOf model.
func (r *Resource) nameOf(m interface{}) (name string) {
	mt := reflect.TypeOf(m)
	mv := reflect.ValueOf(m)
	if mv.IsNil() {
		return
	}
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fv := mv.Field(i)
		switch ft.Name {
		case "Name":
			name = fv.String()
			return
		}
	}
	return
}

//
// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name"`
}

//
// With id and named model.
func (r *Ref) With(id uint, name string) {
	r.ID = id
	r.Name = name
}

//
// TagRef represents a reference to a Tag.
// Contains the tag ID, name, tag source.
type TagRef struct {
	ID     uint   `json:"id" binding:"required"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

//
// With id and named model.
func (r *TagRef) With(id uint, name string, source string) {
	r.ID = id
	r.Name = name
	r.Source = source
}
