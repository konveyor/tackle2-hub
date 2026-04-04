package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// AuthHandler handles auth routes.
type AuthHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AuthHandler) AddRoutes(e *gin.Engine) {
	e.POST(api.AuthAPIKeyRoute, h.CreateKey)
}

// CreateKey godoc
// @summary CreateKey create an API key.
// @description CreateKey create an API key.
// @tags auth
// @produce json
// @success 201 {object} api.APIKey
// @router /auth/apikey [post]
func (h AuthHandler) CreateKey(ctx *gin.Context) {
	r := &APIKey{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	key, err := auth.Hub.UserKey(r.UserId, r.Password, r.Expiration)
	if err != nil {
		h.Respond(ctx,
			http.StatusUnauthorized,
			gin.H{
				"error": err.Error(),
			})
		return
	}
	r.Password = ""
	r.Secret = key.Secret
	h.Respond(ctx, http.StatusCreated, r)
}

// APIKey REST resource.
type APIKey = resource.APIKey

// Required enforces that the user (identified by a token) has
// been granted the necessary scope to access a resource.
func Required(scope string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		rtx := RichContext(ctx)
		token := ctx.GetHeader(Authorization)
		request := &auth.Request{
			Token:  token,
			Scope:  scope,
			Method: ctx.Request.Method,
			DB:     rtx.DB,
		}
		result, err := request.Permit()
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !result.Authenticated {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !result.Authorized {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		rtx.User = result.User
		rtx.Scope.Granted = result.Scopes
		rtx.Scope.Required = append(
			rtx.Scope.Required,
			scope)
	}
}
