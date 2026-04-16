package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// AuthHandler handles auth routes.
type AuthHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AuthHandler) AddRoutes(e *gin.Engine) {
	// Tokens routes
	routeGroup := e.Group("/")
	routeGroup.POST(api.AuthTokensRoute, h.TokenCreate)
	// OIDC routes (hub as provider)
	baseHandler := auth.Hub.Handler()
	strippedHandler := http.StripPrefix(api.OIDCRoutes, baseHandler)
	e.Any(
		api.OIDCRoutes+"/*path",
		func(ctx *gin.Context) {
			path := ctx.Param("path")
			if path == "/login" {
				h.Login(ctx)
			} else {
				strippedHandler.ServeHTTP(ctx.Writer, ctx.Request)
			}
		})
	// IdP routes
	if Settings.Auth.Idp.Enabled {
		idpHandler := auth.Hub.IdpHandler()
		e.GET(api.IdpRoute+"/login", idpHandler.Login)
		e.GET(api.IdpRoute+"/callback", idpHandler.LoginFinished)
	}
	// IdpIdentity routes.
	routeGroup = e.Group("/")
	routeGroup.Use(Required("idp.identities"))
	routeGroup.GET(api.IdpIdentitiesRoute, h.IdpIdentityList)
	routeGroup.GET(api.IdpIdentitiesRoute+"/", h.IdpIdentityList)
	routeGroup.POST(api.IdpIdentitiesRoute, h.IdpIdentityCreate)
	routeGroup.GET(api.IdpIdentityRoute, h.IdpIdentityGet)
	routeGroup.PUT(api.IdpIdentityRoute, h.IdpIdentityUpdate)
	routeGroup.DELETE(api.IdpIdentityRoute, h.IdpIdentityDelete)
	// User routes.
	routeGroup = e.Group("/")
	routeGroup.Use(Required("users"), Transaction)
	routeGroup.GET(api.UsersRoute, h.UserList)
	routeGroup.GET(api.UsersRoute+"/", h.UserList)
	routeGroup.POST(api.UsersRoute, h.UserCreate)
	routeGroup.GET(api.UserRoute, h.UserGet)
	routeGroup.PUT(api.UserRoute, h.UserUpdate)
	routeGroup.DELETE(api.UserRoute, h.UserDelete)
	// Role routes.
	routeGroup = e.Group("/")
	routeGroup.Use(Required("roles"), Transaction)
	routeGroup.GET(api.RolesRoute, h.RoleList)
	routeGroup.GET(api.RolesRoute+"/", h.RoleList)
	routeGroup.POST(api.RolesRoute, h.RoleCreate)
	routeGroup.GET(api.RoleRoute, h.RoleGet)
	routeGroup.PUT(api.RoleRoute, h.RoleUpdate)
	routeGroup.DELETE(api.RoleRoute, h.RoleDelete)
	// Permission routes
	routeGroup = e.Group("/")
	routeGroup.Use(Required("permissions"))
	routeGroup.GET(api.PermissionsRoute, h.PermissionList)
	routeGroup.GET(api.PermissionsRoute+"/", h.PermissionList)
	routeGroup.GET(api.PermissionRoute, h.PermissionGet)
	// Grant routes
	routeGroup = e.Group("/")
	routeGroup.Use(Required("grants"))
	routeGroup.GET(api.AuthGrantsRoute, h.GrantList)
	routeGroup.GET(api.AuthGrantsRoute+"/", h.GrantList)
	routeGroup.GET(api.AuthGrantRoute, h.GrantGet)
	routeGroup.DELETE(api.AuthGrantRoute, h.GrantDelete)
	// Token routes
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tokens"))
	routeGroup.GET(api.AuthTokensRoute, h.TokenList)
	routeGroup.GET(api.AuthTokensRoute+"/", h.TokenList)
	routeGroup.GET(api.AuthTokenRoute, h.TokenGet)
	routeGroup.DELETE(api.AuthTokenRoute, h.TokenDelete)
}

//
// IdpIdentity handlers
//

// IdpIdentityGet godoc
// @summary Get an IDP identity by ID.
// @description Get an IDP identity by ID.
// @tags idpidentities
// @produce json
// @success 200 {object} api.IdpIdentity
// @router /idp/identities/{id} [get]
// @param id path int true "IdpIdentity ID"
func (h AuthHandler) IdpIdentityGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.IdpIdentity{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := IdpIdentity{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// IdpIdentityList godoc
// @summary List all IDP identities.
// @description List all IDP identities.
// @tags idpidentities
// @produce json
// @success 200 {object} []api.IdpIdentity
// @router /idp/identities [get]
func (h AuthHandler) IdpIdentityList(ctx *gin.Context) {
	var list []model.IdpIdentity
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []IdpIdentity{}
	for i := range list {
		r := IdpIdentity{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// IdpIdentityCreate godoc
// @summary Create an IDP identity.
// @description Create an IDP identity.
// @tags idpidentities
// @accept json
// @produce json
// @success 201 {object} api.IdpIdentity
// @router /idp/identities [post]
// @param idpidentity body api.IdpIdentity true "IdpIdentity data"
func (h AuthHandler) IdpIdentityCreate(ctx *gin.Context) {
	r := &IdpIdentity{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	err = h.DB(ctx).Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// IdpIdentityUpdate godoc
// @summary Update an IDP identity.
// @description Update an IDP identity.
// @tags idpidentities
// @accept json
// @success 204
// @router /idp/identities/{id} [put]
// @param id path int true "IdpIdentity ID"
// @param idpidentity body api.IdpIdentity true "IdpIdentity data"
func (h AuthHandler) IdpIdentityUpdate(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &IdpIdentity{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// IdpIdentityDelete godoc
// @summary Delete an IDP identity.
// @description Delete an IDP identity.
// @tags idpidentities
// @success 204
// @router /idp/identities/{id} [delete]
// @param id path int true "IdpIdentity ID"
func (h AuthHandler) IdpIdentityDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.IdpIdentity{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// User handlers
//

// UserGet godoc
// @summary Get a user by ID.
// @description Get a user by ID.
// @tags users
// @produce json
// @success 200 {object} api.User
// @router /users/{id} [get]
// @param id path int true "User ID"
func (h AuthHandler) UserGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.User{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := User{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// UserList godoc
// @summary List all users.
// @description List all users.
// @tags users
// @produce json
// @success 200 {object} []api.User
// @router /users [get]
func (h AuthHandler) UserList(ctx *gin.Context) {
	var list []model.User
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []User{}
	for i := range list {
		m := &list[i]
		r := User{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// UserCreate godoc
// @summary Create a user.
// @description Create a user.
// @tags users
// @accept json
// @produce json
// @success 201 {object} api.User
// @router /users [post]
// @param user body api.User true "User data"
func (h AuthHandler) UserCreate(ctx *gin.Context) {
	r := &User{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.Subject = uuid.New().String()
	m.CreateUser = h.CurrentUser(ctx)
	m.Password = secret.HashPassword(r.Password)
	err = h.DB(ctx).Omit(clause.Associations).Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("Roles").Replace(m.Roles)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	db := h.preLoad(h.DB(ctx), clause.Associations)
	err = db.First(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// UserUpdate godoc
// @summary Update a user.
// @description Update a user.
// @tags users
// @accept json
// @success 204
// @router /users/{id} [put]
// @param id path int true "User ID"
// @param user body api.User true "User data"
func (h AuthHandler) UserUpdate(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &User{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	m.Password = secret.HashPassword(r.Password)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	err = h.DB(ctx).Model(m).Association("Roles").Replace(m.Roles)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// UserDelete godoc
// @summary Delete a user.
// @description Delete a user.
// @tags users
// @success 204
// @router /users/{id} [delete]
// @param id path int true "User ID"
func (h AuthHandler) UserDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.User{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// Role handlers
//

// RoleGet godoc
// @summary Get a role by ID.
// @description Get a role by ID.
// @tags roles
// @produce json
// @success 200 {object} api.Role
// @router /roles/{id} [get]
// @param id path int true "Role ID"
func (h AuthHandler) RoleGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Role{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Role{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// RoleList godoc
// @summary List all roles.
// @description List all roles.
// @tags roles
// @produce json
// @success 200 {object} []api.Role
// @router /roles [get]
func (h AuthHandler) RoleList(ctx *gin.Context) {
	var list []model.Role
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []Role{}
	for i := range list {
		r := Role{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// RoleCreate godoc
// @summary Create a role.
// @description Create a role.
// @tags roles
// @accept json
// @produce json
// @success 201 {object} api.Role
// @router /roles [post]
// @param role body api.Role true "Role data"
func (h AuthHandler) RoleCreate(ctx *gin.Context) {
	r := &Role{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	err = h.DB(ctx).Omit(clause.Associations).Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	err = h.DB(ctx).Model(m).Association("Permissions").Replace(m.Permissions)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	db := h.preLoad(h.DB(ctx), clause.Associations)
	err = db.First(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// RoleUpdate godoc
// @summary Update a role.
// @description Update a role.
// @tags roles
// @accept json
// @success 204
// @router /roles/{id} [put]
// @param id path int true "Role ID"
// @param role body api.Role true "Role data"
func (h AuthHandler) RoleUpdate(ctx *gin.Context) {
	id := h.pk(ctx)
	if id < 1000 {
		h.Status(ctx, http.StatusBadRequest)
		return
	}
	r := &Role{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	err = h.DB(ctx).Model(m).Association("Permissions").Replace(m.Permissions)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// RoleDelete godoc
// @summary Delete a role.
// @description Delete a role.
// @tags roles
// @success 204
// @router /roles/{id} [delete]
// @param id path int true "Role ID"
func (h AuthHandler) RoleDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	if id < 1000 {
		h.Status(ctx, http.StatusBadRequest)
		return
	}
	m := &model.Role{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// Permission handlers
//

// PermissionGet godoc
// @summary Get a permission by ID.
// @description Get a permission by ID.
// @tags permissions
// @produce json
// @success 200 {object} api.Permission
// @router /permissions/{id} [get]
// @param id path int true "Permission ID"
func (h AuthHandler) PermissionGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Permission{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Permission{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// PermissionList godoc
// @summary List all permissions.
// @description List all permissions.
// @tags permissions
// @produce json
// @success 200 {object} []api.Permission
// @router /permissions [get]
func (h AuthHandler) PermissionList(ctx *gin.Context) {
	var list []model.Permission
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []Permission{}
	for i := range list {
		r := Permission{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

//
// Grant handlers
//

// GrantGet godoc
// @summary Get a grant by ID.
// @description Get a grant by ID.
// @tags grants
// @produce json
// @success 200 {object} api.Grant
// @router /auth/grants/{id} [get]
// @param id path int true "Grant ID"
func (h AuthHandler) GrantGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Grant{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Grant{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// GrantList godoc
// @summary List all grants.
// @description List all grants.
// @tags grants
// @produce json
// @success 200 {object} []api.Grant
// @router /auth/grants [get]
func (h AuthHandler) GrantList(ctx *gin.Context) {
	var list []model.Grant
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []Grant{}
	for i := range list {
		r := Grant{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// GrantDelete godoc
// @summary Delete a grant.
// @description Delete a grant.
// @tags grants
// @success 204
// @router /auth/grants/{id} [delete]
// @param id path int true "Grant ID"
func (h AuthHandler) GrantDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Grant{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// Token handlers
//

// TokenCreate godoc
// @summary TokenCreate create a token.
// @description TokenCreate create a token.
// @tags auth
// @produce json
// @success 201 {object} api.Token
// @router /auth/tokens [post]
func (h AuthHandler) TokenCreate(ctx *gin.Context) {
	r := &TokenRequest{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if r.Expiration.IsZero() {
		if r.Lifespan == 0 {
			r.Lifespan = Settings.APIKey.Lifespan
		}
	} else {
		r.Lifespan = int(r.Expiration.Sub(time.Now()) / time.Hour)
	}
	req := auth.TokenRequest{
		Userid:   r.Userid,
		Password: r.Password,
		Lifespan: time.Hour * time.Duration(r.Lifespan),
	}
	token, err := req.Grant()
	if err != nil {
		h.Respond(ctx,
			http.StatusUnauthorized,
			gin.H{
				"error": err.Error(),
			})
		return
	}
	m := &model.Token{}
	m.ID = token.ID
	err = h.DB(ctx).Find(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r2 := Token{}
	r2.With(m)
	r.Token = api.Token(r2)
	r.Secret = token.Secret
	r.Password = "" // redacted

	h.Respond(ctx, http.StatusCreated, r)
}

// TokenGet godoc
// @summary Get a token by ID.
// @description Get a token by ID.
// @tags tokens
// @produce json
// @success 200 {object} api.Token
// @router /auth/tokens/{id} [get]
// @param id path int true "Token ID"
func (h AuthHandler) TokenGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Token{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Token{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// TokenList godoc
// @summary List all tokens.
// @description List all tokens.
// @tags tokens
// @produce json
// @success 200 {object} []api.Token
// @router /auth/tokens [get]
func (h AuthHandler) TokenList(ctx *gin.Context) {
	var list []model.Token
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []Token{}
	for i := range list {
		r := Token{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// TokenDelete godoc
// @summary Delete a token.
// @description Delete a token.
// @tags tokens
// @success 204
// @router /auth/tokens/{id} [delete]
// @param id path int true "Token ID"
func (h AuthHandler) TokenDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Token{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Login OIDC login.
func (h AuthHandler) Login(ctx *gin.Context) {
	authReqID := ctx.Query("authRequestID")
	if authReqID == "" {
		ctx.String(http.StatusBadRequest, "missing authRequestID")
		return
	}
	err := auth.Hub.Login(ctx.Writer, ctx.Request, authReqID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
}

// Auth REST Resources.
type IdpIdentity = resource.IdpIdentity
type User = resource.User
type Role = resource.Role
type Permission = resource.Permission
type Grant = resource.Grant
type Token = resource.Token
type TokenRequest api.TokenRequest

// Required enforces that the user (identified by a token) has
// been granted the necessary scope to access a resource.
// Automatically registers the scope for permission generation.
func Required(scope string) func(*gin.Context) {
	auth.RegisterScope(scope)
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
