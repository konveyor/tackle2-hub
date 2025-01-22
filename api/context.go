package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Response values.
type Response struct {
	Status int
	Body   any
}

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
	// Task manager.
	TaskManager *tasking.Manager
}

// Attach to gin context.
func (r *Context) Attach(ctx *gin.Context) {
	r.Context = ctx
	ctx.Set("RichContext", r)
}

// Detach from gin context
func (r *Context) Detach() {
	delete(r.Context.Keys, "RichContext")
}

// Status sets the values to respond to the request with.
func (r *Context) Status(status int) {
	r.Response = Response{
		Status: status,
		Body:   nil,
	}
}

// Respond sets the values to respond to the request with.
func (r *Context) Respond(status int, body any) {
	r.Response = Response{
		Status: status,
		Body:   body,
	}
}

// RichContext returns a rich context attached to the gin context.
func RichContext(ctx *gin.Context) (rtx *Context) {
	key := "RichContext"
	object, found := ctx.Get(key)
	if !found {
		rtx = &Context{}
		rtx.Attach(ctx)
	} else {
		rtx = object.(*Context)
	}
	rtx.Context = ctx
	return
}

// Transaction handler.
func Transaction(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete:
		rtx := RichContext(ctx)
		err := rtx.DB.Transaction(func(tx *gorm.DB) (err error) {
			db := rtx.DB
			rtx.DB = tx
			ctx.Next()
			rtx.DB = db
			if len(ctx.Errors) > 0 {
				err = ctx.Errors[0]
				ctx.Errors = nil
			}
			return
		})
		if err != nil {
			_ = ctx.Error(err)
		}
	}
}

// Render renders the response based on the Accept: header.
// Opinionated towards json.
func Render() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		rtx := RichContext(ctx)
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
