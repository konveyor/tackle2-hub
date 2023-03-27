package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"gorm.io/gorm"
	"net/http"
)

//
// Context custom settings.
type Context struct {
	*gin.Context
	// DB client.
	DB *gorm.DB
	// User
	User string
	// Scope
	Scopes []auth.Scope
}

//
// WithContext is a rich context.
func WithContext(ctx *gin.Context) (n *Context) {
	key := "RichContext"
	object, found := ctx.Get(key)
	if !found {
		n = &Context{}
		ctx.Set(key, n)
	} else {
		n = object.(*Context)
	}
	n.Context = ctx
	return
}

//
// Transaction handler.
func Transaction(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete:
		rtx := WithContext(ctx)
		err := rtx.DB.Transaction(func(tx *gorm.DB) (err error) {
			db := rtx.DB
			rtx.DB = tx
			ctx.Next()
			rtx.DB = db
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
