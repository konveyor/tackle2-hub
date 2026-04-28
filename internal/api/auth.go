package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"gorm.io/gorm/clause"
)

// AuthHandler handles auth routes.
type AuthHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AuthHandler) AddRoutes(e *gin.Engine) {
	// OIDC routes (hub as provider)
	baseHandler := auth.IdP.Handler()
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
		idpHandler := auth.IdP.IdpHandler()
		e.GET(api.IdpRoute+"/login", idpHandler.Login)
		e.GET(api.IdpRoute+"/callback", idpHandler.LoginFinished)
	}
	// IdpIdentity routes.
	routeGroup := e.Group("/")
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
	routeGroup.POST(api.AuthTokensRoute, h.TokenCreate)
	routeGroup.GET(api.AuthTokensRoute, h.TokenList)
	routeGroup.GET(api.AuthTokensRoute+"/", h.TokenList)
	routeGroup.GET(api.AuthTokenRoute, h.TokenGet)
	routeGroup.DELETE(api.AuthTokenRoute, h.TokenDelete)
	// Device authorization routes with OIDC authentication
	h.setupDeviceAuthRoutes(e)
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

	auth.IdP.Cache().UserSaved((*auth.User)(m))

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

	auth.IdP.Cache().UserSaved((*auth.User)(m))

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

	auth.IdP.Cache().UserDeleted(id)

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

	auth.IdP.Cache().RoleSaved((*auth.Role)(m))

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

	auth.IdP.Cache().RoleSaved((*auth.Role)(m))

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

	auth.IdP.Cache().RoleDeleted(id)

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
// @description TokenCreate create an (apikey) token.
// @tags auth
// @produce json
// @success 201 {object} api.Token
// @router /auth/tokens [post]
func (h AuthHandler) TokenCreate(ctx *gin.Context) {
	r := &PAT{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if r.Lifespan == 0 {
		r.Lifespan = int(Settings.APIKey.Lifespan.Hours())
	}
	if r.Expiration.IsZero() {
		r.Expiration = time.Now().Add(time.Duration(r.Lifespan) * time.Hour)
	}
	subject := h.CurrentUser(ctx)
	lifespan := time.Until(r.Expiration)
	token, err := auth.IdP.NewPAT(subject, lifespan)
	if err != nil {
		h.Respond(ctx,
			http.StatusUnauthorized,
			gin.H{
				"error": err.Error(),
			})
		return
	}

	r.Token = token.Secret

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
	err := auth.IdP.Login(ctx.Writer, ctx.Request, authReqID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
}

// hashKey256 derives a 32-byte key using SHA256.
func hashKey256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// setupDeviceAuthRoutes configures device authorization with OIDC authentication.
func (h AuthHandler) setupDeviceAuthRoutes(e *gin.Engine) {
	dagHandler := auth.IdP.(*auth.Builtin).DagHandler()

	// Lazy-initialized RP client (created on first request)
	var deviceRP rp.RelyingParty
	var cookieHandler *httphelper.CookieHandler

	// Server-side storage for state and PKCE (state -> pkceData mapping)
	// Required because hub acts as both IdP and RP, potentially on different hosts
	type pkceData struct {
		verifier string
		created  time.Time
	}
	stateStore := make(map[string]*pkceData)
	var stateStoreMutex sync.RWMutex

	// ensureRP creates the RP client if not already initialized
	ensureRP := func() (rp.RelyingParty, *httphelper.CookieHandler, error) {
		if deviceRP != nil {
			return deviceRP, cookieHandler, nil
		}

		// Determine issuer URL
		issuer := Settings.Auth.IssuerURL
		if issuer == "" {
			issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
		}

		// Derive proper-sized keys from client secret
		secret := Settings.Auth.Client.Secret
		if secret == "" {
			secret = "default-secret-change-me" // Fallback for development
		}

		// Use SHA256 to derive consistent 32-byte keys
		hashKey := hashKey256([]byte(secret + "-hash"))
		encryptKey := hashKey256([]byte(secret + "-encrypt"))

		// Create cookie handler for session management only (not state/PKCE)
		cookieHandler = httphelper.NewCookieHandler(
			hashKey,
			encryptKey,
			httphelper.WithUnsecure(),
			httphelper.WithSameSite(http.SameSiteLaxMode),
		)

		// Create OIDC RP client without cookie handler for state
		// We'll handle state and PKCE manually with server-side storage
		var err error
		deviceRP, err = rp.NewRelyingPartyOIDC(
			context.Background(),
			issuer,
			"device-verifier",
			Settings.Auth.Client.Secret,
			Settings.Addon.Hub.URL+api.AuthDevAuthCallback,
			[]string{"openid"},
		)
		return deviceRP, cookieHandler, err
	}

	// Login redirect handler - manual state and PKCE with server-side storage
	loginHandler := func(ctx *gin.Context) {
		rpClient, _, err := ensureRP()
		if err != nil {
			Log.Error(err, "Failed to initialize device verification RP client")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Generate state
		state := uuid.New().String()

		// Generate PKCE code verifier (43-128 characters)
		verifierBytes := make([]byte, 32)
		_, err = rand.Read(verifierBytes)
		if err != nil {
			Log.Error(err, "Failed to generate PKCE verifier")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

		// Generate PKCE code challenge (SHA256 hash of verifier)
		hash := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

		// Store state and verifier server-side
		stateStoreMutex.Lock()
		stateStore[state] = &pkceData{
			verifier: codeVerifier,
			created:  time.Now(),
		}
		// Clean up old states (>10 minutes old)
		now := time.Now()
		for s, pd := range stateStore {
			if now.Sub(pd.created) > 10*time.Minute {
				delete(stateStore, s)
			}
		}
		stateStoreMutex.Unlock()

		// Build authorize URL with state and PKCE challenge
		authURL := rp.AuthURL(state, rpClient, rp.WithCodeChallenge(codeChallenge))

		http.Redirect(ctx.Writer, ctx.Request, authURL, http.StatusFound)
	}

	// Login handler - initiates OAuth flow with state management
	e.GET(api.AuthDevAuthRoute+"/login", loginHandler)

	// Callback handler - manual code exchange with server-side PKCE
	e.GET(api.AuthDevAuthCallback, func(ctx *gin.Context) {
		rpClient, cookies, err := ensureRP()
		if err != nil {
			Log.Error(err, "Failed to initialize device verification RP client")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Get state and code from query parameters
		state := ctx.Query("state")
		code := ctx.Query("code")

		// Retrieve and validate state from server-side storage
		stateStoreMutex.Lock()
		pkceInfo, found := stateStore[state]
		if found {
			delete(stateStore, state) // Use once and delete
		}
		stateStoreMutex.Unlock()

		if !found {
			Log.Error(nil, "Invalid state in callback", "state", state)
			http.Error(ctx.Writer, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// Exchange code for tokens with PKCE verifier
		tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](
			ctx.Request.Context(),
			code,
			rpClient,
			rp.WithCodeVerifier(pkceInfo.verifier),
		)
		if err != nil {
			Log.Error(err, "Code exchange failed")
			http.Error(ctx.Writer, "Failed to exchange code for token", http.StatusInternalServerError)
			return
		}

		subject := tokens.IDTokenClaims.Subject

		// Store only the subject in the cookie (not the full tokens)
		// This avoids cookie size limits and we only need the subject for device auth
		err = cookies.SetCookie(ctx.Writer, "oidc_subject", subject)
		if err != nil {
			Log.Error(err, "Failed to set session cookie")
			http.Error(ctx.Writer, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Redirect to device authorization page
		http.Redirect(ctx.Writer, ctx.Request, api.AuthDevAuthRoute, http.StatusFound)
	})

	// Auth middleware - checks for valid OIDC session
	authMiddleware := func(ctx *gin.Context) {
		_, cookies, err := ensureRP()
		if err != nil {
			Log.Error(err, "Failed to initialize device verification RP client")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Check for session cookie containing subject
		subject, err := cookies.CheckCookie(ctx.Request, "oidc_subject")
		if err != nil || subject == "" {
			// No session - redirect to login handler
			ctx.Redirect(http.StatusFound, api.AuthDevAuthRoute+"/login")
			ctx.Abort()
			return
		}

		// Store user subject in context for DagHandler
		ctx.Set("oidc_subject", subject)
		ctx.Next()
	}

	// Protected device authorization routes
	deviceGroup := e.Group(api.AuthDevAuthRoute)
	deviceGroup.Use(authMiddleware)
	deviceGroup.GET("", dagHandler.Verify)
	deviceGroup.POST("", dagHandler.VerifySubmit)
}

// Auth REST Resources.
type IdpIdentity = resource.IdpIdentity
type User = resource.User
type Role = resource.Role
type Permission = resource.Permission
type Grant = resource.Grant
type Token = resource.Token
type PAT api.PAT

// Required enforces that the user (identified by a token) has
// been granted the necessary scope to access a resource.
// Automatically registers the scope for permission generation.
func Required(scope string) func(*gin.Context) {
	auth.RegisterScope(scope)
	return func(ctx *gin.Context) {
		rtx := RichContext(ctx)
		header := ctx.GetHeader(Authorization)
		request := &auth.Request{
			Scope:  scope,
			Method: ctx.Request.Method,
			DB:     rtx.DB,
			CTX:    ctx,
		}
		request.With(header)
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
