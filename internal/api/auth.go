package api

import (
	"net/http"

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
	// APIKey routes
	routeGroup := e.Group("/")
	routeGroup.Use(Required("apikeys"))
	routeGroup.POST(api.AuthAPIKeyRoute, h.CreateKey)
	routeGroup.GET(api.AuthAPIKeysRoute, h.APIKeyList)
	routeGroup.GET(api.AuthAPIKeysRoute+"/", h.APIKeyList)
	routeGroup.GET(api.AuthAPIKeyIDRoute, h.APIKeyGet)
	routeGroup.DELETE(api.AuthAPIKeyIDRoute, h.APIKeyDelete)
	// OIDC routes.
	h2 := auth.Hub.Handler()
	h2 = http.StripPrefix(api.OIDCRoutes, h2)
	routeGroup = e.Group(api.OIDCRoutes)
	routeGroup.Any("/*path", gin.WrapH(h2))
	// IdpIdentity routes
	routeGroup = e.Group("/")
	routeGroup.Use(Required("idp.identities"))
	routeGroup.GET(api.IdpIdentitiesRoute, h.IdpIdentityList)
	routeGroup.GET(api.IdpIdentitiesRoute+"/", h.IdpIdentityList)
	routeGroup.POST(api.IdpIdentitiesRoute, h.IdpIdentityCreate)
	routeGroup.GET(api.IdpIdentityRoute, h.IdpIdentityGet)
	routeGroup.PUT(api.IdpIdentityRoute, h.IdpIdentityUpdate)
	routeGroup.DELETE(api.IdpIdentityRoute, h.IdpIdentityDelete)
	// User routes
	routeGroup = e.Group("/")
	routeGroup.Use(Required("users"), Transaction)
	routeGroup.GET(api.UsersRoute, h.UserList)
	routeGroup.GET(api.UsersRoute+"/", h.UserList)
	routeGroup.POST(api.UsersRoute, h.UserCreate)
	routeGroup.GET(api.UserRoute, h.UserGet)
	routeGroup.PUT(api.UserRoute, h.UserUpdate)
	routeGroup.DELETE(api.UserRoute, h.UserDelete)
	// Role routes
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
	routeGroup.POST(api.PermissionsRoute, h.PermissionCreate)
	routeGroup.GET(api.PermissionRoute, h.PermissionGet)
	routeGroup.PUT(api.PermissionRoute, h.PermissionUpdate)
	routeGroup.DELETE(api.PermissionRoute, h.PermissionDelete)
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
	kr := auth.KeyRequest{
		Userid:   r.Userid,
		Password: r.Password,
		Lifespan: r.Lifespan,
	}
	key, err := kr.Grant()
	if err != nil {
		h.Respond(ctx,
			http.StatusUnauthorized,
			gin.H{
				"error": err.Error(),
			})
		return
	}
	r.Userid = ""         // redacted.
	r.Password = ""       // redacted.
	r.Secret = key.Secret // Plain text (user must save it).
	r.Digest = key.Digest // Hash (for reference).
	h.Respond(ctx, http.StatusCreated, r)
}

// APIKeyGet godoc
// @summary Get an API key by ID.
// @description Get an API key by ID.
// @tags apikeys
// @produce json
// @success 200 {object} api.APIKey
// @router /auth/apikeys/{id} [get]
// @param id path int true "APIKey ID"
func (h AuthHandler) APIKeyGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.APIKey{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := APIKey{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// APIKeyList godoc
// @summary List all API keys.
// @description List all API keys.
// @tags apikeys
// @produce json
// @success 200 {object} []api.APIKey
// @router /auth/apikeys [get]
func (h AuthHandler) APIKeyList(ctx *gin.Context) {
	var list []model.APIKey
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []APIKey{}
	for i := range list {
		r := APIKey{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// APIKeyDelete godoc
// @summary Delete an API key.
// @description Delete an API key.
// @tags apikeys
// @success 204
// @router /auth/apikeys/{id} [delete]
// @param id path int true "APIKey ID"
func (h AuthHandler) APIKeyDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.APIKey{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = auth.Hub.Delete(m.Digest)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
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
// @router /idpidentities/{id} [get]
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
// @router /idpidentities [get]
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
// @router /idpidentities [post]
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
// @router /idpidentities/{id} [put]
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
// @router /idpidentities/{id} [delete]
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

// PermissionCreate godoc
// @summary Create a permission.
// @description Create a permission.
// @tags permissions
// @accept json
// @produce json
// @success 201 {object} api.Permission
// @router /permissions [post]
// @param permission body api.Permission true "Permission data"
func (h AuthHandler) PermissionCreate(ctx *gin.Context) {
	r := &Permission{}
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

// PermissionUpdate godoc
// @summary Update a permission.
// @description Update a permission.
// @tags permissions
// @accept json
// @success 204
// @router /permissions/{id} [put]
// @param id path int true "Permission ID"
// @param permission body api.Permission true "Permission data"
func (h AuthHandler) PermissionUpdate(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Permission{}
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

// PermissionDelete godoc
// @summary Delete a permission.
// @description Delete a permission.
// @tags permissions
// @success 204
// @router /permissions/{id} [delete]
// @param id path int true "Permission ID"
func (h AuthHandler) PermissionDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Permission{}
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

// Auth REST Resources.
type APIKey = resource.APIKey
type IdpIdentity = resource.IdpIdentity
type User = resource.User
type Role = resource.Role
type Permission = resource.Permission

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
