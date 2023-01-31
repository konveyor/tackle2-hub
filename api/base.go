package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

var Log = logging.WithName("api")

//
// BaseHandler base handler.
type BaseHandler struct {
	// DB
	DB *gorm.DB
	// k8s Client
	Client client.Client
}

// With database and k8s client.
func (h *BaseHandler) With(db *gorm.DB, client client.Client) {
	h.DB = db.Debug()
	h.Client = client
}

//
// reportError reports hub errors as http statuses.
func (h *BaseHandler) reportError(ctx *gin.Context, err error) {
	// gin binding errors
	if errors.Is(err, validator.ValidationErrors{}) {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			})
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if ctx.Request.Method == http.MethodDelete {
			ctx.Status(http.StatusNoContent)
			return
		}
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"error": err.Error(),
			})
		return
	}

	if errors.Is(err, os.ErrNotExist) {
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"error": err.Error(),
			})
		return
	}

	if errors.Is(err, model.DependencyCyclicError{}) {
		ctx.JSON(
			http.StatusConflict,
			gin.H{
				"error": err.Error(),
			})
		return
	}

	sqliteErr := &sqlite3.Error{}
	if errors.As(err, sqliteErr) {
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintUnique,
			sqlite3.ErrConstraintPrimaryKey:
			ctx.JSON(
				http.StatusConflict,
				gin.H{
					"error": err.Error(),
				})
			return
		}
	}

	ctx.JSON(
		http.StatusInternalServerError,
		gin.H{
			"error": err.Error(),
		})

	url := ctx.Request.URL.String()
	log.Error(
		err,
		"Request failed.",
		"method",
		ctx.Request.Method,
		"url",
		url)
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
	isExported := func(ft reflect.StructField) bool {
		return ft.PkgPath == ""
	}
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
			if !isExported(ft) {
				continue
			}
			switch fv.Kind() {
			case reflect.Struct:
				if ft.Anonymous {
					inspect(fv.Addr().Interface())
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
