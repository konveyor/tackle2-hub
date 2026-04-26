package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//
// IdpHandler
//

// IdpHandler handles external IdP federation (hub as relying party).
type IdpHandler struct {
	rpClient rp.RelyingParty
	db       *gorm.DB
	storage  *Storage
	cache    *Cache
}

// Login initiates the external IdP authentication flow.
func (h *IdpHandler) Login(ctx *gin.Context) {
	if !Settings.Auth.Idp.Enabled {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	flow := &IdpLogin{handler: h, ctx: ctx}
	flow.begin()
}

// LoginFinished handles the redirect back from the external IdP.
func (h *IdpHandler) LoginFinished(ctx *gin.Context) {
	if !Settings.Auth.Idp.Enabled {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	flow := &IdpLogin{handler: h, ctx: ctx}
	flow.complete()
}

// parseAccessToken parses the access token JWT and returns the claims.
func (h *IdpHandler) parseAccessToken(accessToken string) (claims map[string]any, err error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	mapClaims, cast := token.Claims.(jwt.MapClaims)
	if !cast {
		err = liberr.New("invalid token claims type")
		return
	}

	claims = make(map[string]any)
	for k, v := range mapClaims {
		claims[k] = v
	}

	return
}

// identitySaved notification the Idp identity has been saved.
func (h *IdpHandler) identitySaved(m *Identity) {
	h.cache.IdentitySaved(m)
}

//
// IdpLogin
//

// IdpLogin represents the state of an OIDC login flow with an external IdP.
type IdpLogin struct {
	handler *IdpHandler
	ctx     *gin.Context

	state         string
	codeVerifier  string
	authRequestID string
	code          string

	tokens            *oidc.Tokens[*oidc.IDTokenClaims]
	userInfo          *oidc.UserInfo
	accessTokenClaims map[string]any
	identity          *Identity
}

// begin initiates the external IdP authentication flow.
func (f *IdpLogin) begin() {
	// Get authRequestID from query parameter (if present)
	f.authRequestID = f.ctx.Query("authRequestID")

	// Generate state and code verifier for PKCE
	f.state = f.genState()
	f.codeVerifier = f.genState()

	// Build authorization URL with PKCE
	authURL := rp.AuthURL(
		f.state,
		f.handler.rpClient,
		rp.WithCodeChallenge(oidc.NewSHACodeChallenge(f.codeVerifier)),
	)

	// Store state and code verifier in cookies for validation in callback
	f.ctx.SetCookie("idp_state", f.state, 600, "/", "", false, true)
	f.ctx.SetSameSite(http.SameSiteLaxMode)

	f.ctx.SetCookie("idp_verifier", f.codeVerifier, 600, "/", "", false, true)
	f.ctx.SetSameSite(http.SameSiteLaxMode)

	// Store authRequestID if present (for completing OIDC flow)
	if f.authRequestID != "" {
		f.ctx.SetCookie("idp_auth_request", f.authRequestID, 600, "/", "", false, true)
		f.ctx.SetSameSite(http.SameSiteLaxMode)
	}

	// Redirect to external IdP
	f.ctx.Redirect(http.StatusFound, authURL)
}

// complete handles the redirect back from the external IdP.
func (f *IdpLogin) complete() {
	err := f.validate()
	if err != nil {
		_ = f.ctx.Error(err)
		return
	}
	err = f.exchangeCode()
	if err != nil {
		_ = f.ctx.Error(err)
		return
	}
	err = f.fetchUserInfo(f.ctx.Request.Context())
	if err != nil {
		_ = f.ctx.Error(err)
		return
	}
	err = f.parseAccessToken()
	if err != nil {
		_ = f.ctx.Error(err)
		return
	}
	err = f.ensureIdentity()
	if err != nil {
		_ = f.ctx.Error(err)
		return
	}
	err = f.issueTokens()
	if err != nil {
		_ = f.ctx.Error(err)
		return
	}
}

// validate validates the callback request and extracts flow state.
func (f *IdpLogin) validate() (err error) {
	// Validate state (CSRF protection)
	state := f.ctx.Query("state")
	stateCookie, err := f.ctx.Cookie("idp_state")
	if err != nil || state != stateCookie {
		err = &BadRequestError{Reason: "invalid state parameter"}
		return
	}
	f.state = state

	// Get code verifier from cookie
	verifierCookie, err := f.ctx.Cookie("idp_verifier")
	if err != nil {
		err = &BadRequestError{Reason: "missing code verifier"}
		return
	}
	f.codeVerifier = verifierCookie

	// Get authRequestID from cookie (if present)
	if authReqCookie, err := f.ctx.Cookie("idp_auth_request"); err == nil {
		f.authRequestID = authReqCookie
	}

	// Clear cookies
	f.ctx.SetCookie("idp_state", "", -1, "/", "", false, true)
	f.ctx.SetCookie("idp_verifier", "", -1, "/", "", false, true)
	f.ctx.SetCookie("idp_auth_request", "", -1, "/", "", false, true)

	// Check for error from IdP
	if errParam := f.ctx.Query("error"); errParam != "" {
		errDesc := f.ctx.Query("error_description")
		err = &BadRequestError{Reason: fmt.Sprintf("IdP error: %s - %s", errParam, errDesc)}
		return
	}

	// Get authorization code
	f.code = f.ctx.Query("code")
	if f.code == "" {
		err = &BadRequestError{Reason: "missing authorization code"}
		return
	}

	return
}

// exchangeCode exchanges the authorization code for tokens.
func (f *IdpLogin) exchangeCode() (err error) {
	f.tokens, err = rp.CodeExchange[*oidc.IDTokenClaims](
		f.ctx.Request.Context(),
		f.code,
		f.handler.rpClient,
		rp.WithCodeVerifier(f.codeVerifier),
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// fetchUserInfo fetches user info from the IdP.
func (f *IdpLogin) fetchUserInfo(ctx context.Context) (err error) {
	f.userInfo, err = rp.Userinfo[*oidc.UserInfo](
		ctx,
		f.tokens.AccessToken,
		f.tokens.TokenType,
		f.tokens.IDTokenClaims.GetSubject(),
		f.handler.rpClient,
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// parseAccessToken parses the access token to extract claims.
func (f *IdpLogin) parseAccessToken() (err error) {
	f.accessTokenClaims, err = f.handler.parseAccessToken(f.tokens.AccessToken)
	return
}

// ensureIdentity ensures the identity created/updated.
func (f *IdpLogin) ensureIdentity() (err error) {
	f.identity = f.buildIdentity()
	db := f.handler.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "subject"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"refreshToken",
			"expiration",
			"lastAuthenticated",
			"lastRefreshed",
			"scopes",
			"roles",
			"userId",
			"email",
			"updateUser",
		}),
	})
	err = db.Create(f.identity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	f.handler.identitySaved(f.identity)
	return
}

// buildIdentity builds an IdpIdentity from userinfo and tokens.
func (f *IdpLogin) buildIdentity() (identity *Identity) {
	email := f.userInfo.Email
	if email == "" {
		email = fmt.Sprintf("%s@unknown", f.userInfo.Subject)
	}

	userid := f.userInfo.PreferredUsername
	if userid == "" {
		userid = f.userInfo.Subject
	}

	scopes, roles := f.extractClaims()

	expiration := time.Now().Add(Settings.Auth.Token.RefreshLifespan)

	if f.tokens.Expiry.After(time.Now()) {
		expiration = f.tokens.Expiry
	}

	identity = &Identity{
		Issuer:            Settings.Auth.Idp.Name,
		Subject:           f.userInfo.Subject,
		Userid:            userid,
		Email:             email,
		RefreshToken:      f.tokens.RefreshToken,
		Expiration:        expiration,
		LastAuthenticated: time.Now(),
		LastRefreshed:     time.Now(),
		Scopes:            scopes,
		Roles:             roles,
	}
	identity.CreateUser = "system"
	identity.UpdateUser = "system"

	return
}

// extractClaims extracts scopes and roles from access token claims.
func (f *IdpLogin) extractClaims() (scopes string, roles string) {
	if f.accessTokenClaims == nil {
		return
	}

	// Extract scopes
	if scopeClaim, found := f.accessTokenClaims["scope"]; found {
		scopes = f.asString(scopeClaim)
	}

	// Extract roles
	if roleClaim, found := f.accessTokenClaims["roles"]; found {
		roles = f.asString(roleClaim)
	}

	return
}

// issueTokens creates a hub OIDC session and redirects to UI.
func (f *IdpLogin) issueTokens() (err error) {
	var redirectURI string
	var state string
	var requestID string

	// Use external IdP subject directly
	subject := f.identity.Subject

	if f.authRequestID != "" {
		// Complete existing OIDC authorization request
		var authReq op.AuthRequest
		authReq, err = f.handler.storage.AuthRequestByID(f.ctx.Request.Context(), f.authRequestID)
		if err != nil {
			return
		}

		// Update the auth request with the authenticated user
		if ar, cast := authReq.(*AuthRequest); cast {
			ar.subject = subject
		}

		redirectURI = authReq.GetRedirectURI()
		state = authReq.GetState()
		requestID = f.authRequestID
	} else {
		// Create a new authorization request (standalone IdP login)
		var authReq *oidc.AuthRequest
		var req op.AuthRequest
		authReq = &oidc.AuthRequest{
			ClientID:     Settings.Auth.Client.ID,
			RedirectURI:  Settings.Auth.Client.RedirectURIs[0],
			Scopes:       []string{"openid", "profile", "email"},
			ResponseType: oidc.ResponseTypeCode,
			State:        f.genState(),
		}

		req, err = f.handler.storage.CreateAuthRequest(f.ctx.Request.Context(), authReq, subject)
		if err != nil {
			return
		}

		redirectURI = authReq.RedirectURI
		state = authReq.State
		requestID = req.GetID()
	}

	// Generate authorization code
	code := f.genState()
	err = f.handler.storage.SaveAuthCode(f.ctx.Request.Context(), requestID, code)
	if err != nil {
		return
	}

	// Build redirect URL with code
	redirectURL := fmt.Sprintf(
		"%s?code=%s&state=%s",
		redirectURI,
		code,
		state,
	)

	// Redirect to UI
	f.ctx.Redirect(http.StatusFound, redirectURL)
	return
}

// asString converts a claim value to space-separated string.
func (f *IdpLogin) asString(claim any) (s string) {
	var values []string
	switch v := claim.(type) {
	case []interface{}:
		for _, item := range v {
			if s, cast := item.(string); cast {
				values = append(values, s)
			}
		}
	case []string:
		values = v
	case string:
		// Already space-separated or single value
		return v
	default:
		Log.Info(
			"Unexpected claim type",
			"type",
			fmt.Sprintf("%T", claim))
		return
	}
	s = strings.Join(values, " ")
	return
}

// RefreshIdentity refreshes an IdpIdentity using its refresh token.
func (f *IdpLogin) RefreshIdentity() (err error) {
	f.tokens, err = rp.RefreshTokens[*oidc.IDTokenClaims](
		context.Background(),
		f.handler.rpClient,
		f.identity.RefreshToken,
		"",
		"",
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = f.fetchUserInfo(context.Background())
	if err != nil {
		return
	}
	err = f.parseAccessToken()
	if err != nil {
		return
	}
	f.identity = f.buildIdentity()
	err = f.ensureIdentity()
	return
}

// genState generates a random state string.
func (f *IdpLogin) genState() (state string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	state = base64.URLEncoding.EncodeToString(b)
	return
}
