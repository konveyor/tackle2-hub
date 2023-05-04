package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"gorm.io/gorm"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// k8s Client
	Client client.Client
	// Response
	Response Response
}

//
// Response values.
type Response struct {
	Status int
	Body   interface{}
}

//
// Status sets the values to respond to the request with.
func (r *Context) Status(status int) {
	r.Response = Response{
		Status: status,
		Body:   nil,
	}
}

//
// Respond sets the values to respond to the request with.
func (r *Context) Respond(status int, body interface{}) {
	r.Response = Response{
		Status: status,
		Body:   body,
	}
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

//
// Render renders the response based on the Accept: header.
// Opinionated towards json.
func Render() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		rtx := WithContext(ctx)
		if rtx.Response.Body != nil {
			ctx.Negotiate(
				rtx.Response.Status,
				gin.Negotiate{
					Offered: BindMIMEs,
					Data:    rtx.Response.Body})
			return
		}
		ctx.Status(rtx.Response.Status)
	}
}
