package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"net/http"
)

//
// Routes
const (
	AuthRoot      = "/auth"
	AuthLoginRoot = AuthRoot + "/login"
)

//
// AuthHandler handles auth routes.
type AuthHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h AuthHandler) AddRoutes(e *gin.Engine) {
	e.POST(AuthLoginRoot, h.Login)
}

// Login godoc
// @summary Login and obtain a bearer token.
// @description Login and obtain a bearer token.
// @tags post
// @produce json
// @success 201 {string} token
// @router /auth/login [post]
func (h AuthHandler) Login(ctx *gin.Context) {
	token := ""
	r := &Login{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.reportError(ctx, err)
		return
	}
	token, err = auth.Remote.Login(r.User, r.Password)
	if err != nil {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": err.Error(),
			})
		return
	}
	ctx.JSON(http.StatusCreated, token)
}

//
// Login REST resource.
type Login struct {
	User     string `json:"user"`
	Password string `json:"password"`
}
