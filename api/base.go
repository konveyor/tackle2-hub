package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"net/http"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

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
			if !fv.CanSet() {
				continue
			}
			switch fv.Kind() {
			case reflect.Struct:
				inspect(fv.Addr().Interface())
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
// REST resource.
type Resource struct {
	ID         uint      `json:"id"`
	CreateUser string    `json:"createUser"`
	UpdateUser string    `json:"updateUser"`
	CreateTime time.Time `json:"createTime"`
}

//
// Update the resource with the model.
func (r *Resource) With(m *model.Model) {
	r.ID = m.ID
	r.CreateUser = m.CreateUser
	r.UpdateUser = m.UpdateUser
	r.CreateTime = m.CreateTime
}

//
// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name"`
}

func (r *Ref) With(id uint, name string) {
	r.ID = id
	r.Name = name
}
