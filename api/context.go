package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

const (
	// CtxDB is the DB client stored on the context.
	CtxDB = "DB"
)

//
// SetDB inject the DB into the context.
func SetDB(ctx *gin.Context, db *gorm.DB) {
	ctx.Set(CtxDB, db)
}

//
// GetDB extract DB from the context.
func GetDB(ctx *gin.Context) (db *gorm.DB) {
	object, _ := ctx.Get(CtxDB)
	db = object.(*gorm.DB)
	return
}

//
// Transaction handler.
func Transaction(ctx *gin.Context) {
	db := GetDB(ctx)
	switch ctx.Request.Method {
	case http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete:
		err := db.Transaction(func(tx *gorm.DB) (err error) {
			ctx.Set(CtxDB, tx)
			SetDB(ctx, tx)
			ctx.Next()
			SetDB(ctx, db)
			if len(ctx.Errors) > 0 {
				err = ctx.Errors[0]
			}
			return
		})
		if err != nil {
			_ = ctx.Error(err)
		}
	}
}
