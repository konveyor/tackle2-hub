package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//
// BaseHandler base handler.
type BaseHandler struct {
	// DB
	DB *gorm.DB
	// k8s Client
	Client client.Client
	// Auth provider
	AuthProvider auth.Provider
}

// With database and k8s client.
func (h *BaseHandler) With(db *gorm.DB, client client.Client, provider auth.Provider) {
	h.DB = db.Debug()
	h.Client = client
	h.AuthProvider = provider
}

//
// getFailed handles Get() errors.
func (h *BaseHandler) getFailed(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
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
	ctx.JSON(
		http.StatusInternalServerError,
		gin.H{
			"error": err.Error(),
		})

	url := ctx.Request.URL.String()
	log.Error(
		err,
		"Get failed.",
		"url",
		url)
}

//
// listFailed handles List() errors.
func (h *BaseHandler) listFailed(ctx *gin.Context, err error) {
	ctx.JSON(
		http.StatusInternalServerError,
		gin.H{
			"error": err.Error(),
		})

	url := ctx.Request.URL.String()
	log.Error(
		err,
		"List failed.",
		"url",
		url)
}

//
// createFailed handles Create() errors.
func (h *BaseHandler) createFailed(ctx *gin.Context, err error) {
	status := http.StatusInternalServerError
	sqliteErr := &sqlite3.Error{}

	if errors.As(err, sqliteErr) {
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintUnique,
			sqlite3.ErrConstraintPrimaryKey:
			status = http.StatusConflict
		}
	}

	ctx.JSON(
		status,
		gin.H{
			"error": err.Error(),
		})

	url := ctx.Request.URL.String()
	log.Error(
		err,
		"Create failed.",
		"url",
		url)
}

//
// updateFailed handles Update() errors.
func (h *BaseHandler) updateFailed(ctx *gin.Context, err error) {
	status := http.StatusInternalServerError
	sqliteErr := &sqlite3.Error{}

	if errors.As(err, sqliteErr) {
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintUnique:
			status = http.StatusConflict
		}
	}

	ctx.JSON(
		status,
		gin.H{
			"error": err.Error(),
		})

	url := ctx.Request.URL.String()
	log.Error(
		err,
		"Update failed.",
		"url",
		url)
}

//
// deleteFailed handles Delete() errors.
func (h *BaseHandler) deleteFailed(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Status(http.StatusOK)
		return
	}
	ctx.JSON(
		http.StatusInternalServerError,
		gin.H{
			"error": err.Error(),
		})

	url := ctx.Request.URL.String()
	log.Error(
		err,
		"Delete failed.",
		"url",
		url)
}

//
// bindFailed handles errors from BindJSON().
func (h *BaseHandler) bindFailed(ctx *gin.Context, err error) {
	ctx.JSON(
		http.StatusBadRequest,
		gin.H{
			"error": err.Error(),
		})
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
	ctx.Request.Body = ioutil.NopCloser(bfr)
	return
}

//tady dát volání auth providera pro username možná s cachováním
//
// Get user info from Keycloak
func (h *BaseHandler) CurrentUsername(ctx *gin.Context) (username string) {
	fmt.Printf("++++++++++++++ gin ctx: %v", ctx)
	token := ctx.GetHeader("Authorization")

	username, err := h.AuthProvider.GetUsername(token)
	if err != nil {
		fmt.Printf("+++++++++++++++ failed get userInfo, err: %v", err)
		return ""
	}

	fmt.Printf("+++++++++++++++++++++ userName: %v", username)

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
