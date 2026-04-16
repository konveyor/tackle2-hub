package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"gorm.io/gorm"
)

//
// IdpHandler
//

// IdpHandler handles external IdP federation (hub as relying party).
type IdpHandler struct {
	rpClient rp.RelyingParty
	db       *gorm.DB
	storage  *Storage
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

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = liberr.New("invalid token claims type")
		return
	}

	claims = make(map[string]any)
	for k, v := range mapClaims {
		claims[k] = v
	}

	return
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

	idpIdentity *model.IdpIdentity
}

// begin initiates the external IdP authentication flow.
func (f *IdpLogin) begin() {
	// Get authRequestID from query parameter (if present)
	f.authRequestID = f.ctx.Query("authRequestID")

	// Generate state and code verifier for PKCE
	f.state = f.generateState()
	f.codeVerifier = f.generateState()

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

	err = f.fetchUserInfo()
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

	Log.Info("External IdP authentication successful", "subject", f.idpIdentity.Subject)

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
func (f *IdpLogin) fetchUserInfo() (err error) {
	f.userInfo, err = rp.Userinfo[*oidc.UserInfo](
		f.ctx.Request.Context(),
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
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// ensureIdentity finds existing identity or creates new one from IdP userinfo.
func (f *IdpLogin) ensureIdentity() (err error) {
	f.idpIdentity = &model.IdpIdentity{}
	err = f.handler.db.First(f.idpIdentity, "subject", f.userInfo.Subject).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = f.createIdentity()
		if err != nil {
			return
		}
	} else if err != nil {
		err = liberr.Wrap(err)
		return
	} else {
		err = f.updateIdentity()
		if err != nil {
			return
		}
	}
	return
}

// createIdentity creates a new IdpIdentity for first-time login.
func (f *IdpLogin) createIdentity() (err error) {
	// Extract user details from userinfo
	email := f.userInfo.Email
	if email == "" {
		email = fmt.Sprintf("%s@external", f.userInfo.Subject)
	}

	userid := f.userInfo.PreferredUsername
	if userid == "" {
		userid = f.userInfo.Subject
	}

	// Extract scopes and roles from access token
	scopes, roles := f.extractClaims()

	// Create IdpIdentity
	expiration := time.Now().Add(
		time.Duration(Settings.Auth.Token.RefreshLifespan) * time.Second,
	)
	if f.tokens.Expiry.After(time.Now()) {
		expiration = f.tokens.Expiry
	}

	f.idpIdentity = &model.IdpIdentity{
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
	f.idpIdentity.CreateUser = "system"

	err = f.handler.db.Create(f.idpIdentity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// updateIdentity updates an existing IdpIdentity with new tokens and userinfo.
func (f *IdpLogin) updateIdentity() (err error) {
	// Update user details from userinfo
	if f.userInfo.Email != "" {
		f.idpIdentity.Email = f.userInfo.Email
	}
	if f.userInfo.PreferredUsername != "" {
		f.idpIdentity.Userid = f.userInfo.PreferredUsername
	}

	// Extract scopes and roles from access token
	scopes, roles := f.extractClaims()

	// Update tokens, timestamps, and claims
	f.idpIdentity.RefreshToken = f.tokens.RefreshToken
	if f.tokens.Expiry.After(time.Now()) {
		f.idpIdentity.Expiration = f.tokens.Expiry
	}
	f.idpIdentity.LastAuthenticated = time.Now()
	f.idpIdentity.LastRefreshed = time.Now()
	f.idpIdentity.Scopes = scopes
	f.idpIdentity.Roles = roles
	f.idpIdentity.UpdateUser = "system"

	err = f.handler.db.Save(f.idpIdentity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// extractClaims extracts scopes and roles from access token claims.
func (f *IdpLogin) extractClaims() (scopes string, roles string) {
	if f.accessTokenClaims == nil {
		return
	}

	// Extract scopes
	if scopeClaim, found := f.accessTokenClaims["scope"]; found {
		scopes = f.stringFromClaim(scopeClaim)
	}

	// Extract roles
	if roleClaim, found := f.accessTokenClaims["roles"]; found {
		roles = f.stringFromClaim(roleClaim)
	}

	return
}

// issueTokens creates a hub OIDC session and redirects to UI.
func (f *IdpLogin) issueTokens() (err error) {
	var redirectURI string
	var state string
	var requestID string

	// Use external IdP subject directly
	subject := f.idpIdentity.Subject

	if f.authRequestID != "" {
		// Complete existing OIDC authorization request
		authReq, err := f.handler.storage.AuthRequestByID(f.ctx.Request.Context(), f.authRequestID)
		if err != nil {
			err = liberr.Wrap(err)
			return err
		}

		// Update the auth request with the authenticated user
		if ar, ok := authReq.(*AuthRequest); ok {
			ar.subject = subject
		}

		redirectURI = authReq.GetRedirectURI()
		state = authReq.GetState()
		requestID = f.authRequestID
	} else {
		// Create a new authorization request (standalone IdP login)
		authReq := &oidc.AuthRequest{
			ClientID:     Settings.Auth.Client.ID,
			RedirectURI:  Settings.Auth.Client.RedirectURIs[0],
			Scopes:       []string{"openid", "profile", "email"},
			ResponseType: oidc.ResponseTypeCode,
			State:        f.generateState(),
		}

		req, err := f.handler.storage.CreateAuthRequest(f.ctx.Request.Context(), authReq, subject)
		if err != nil {
			err = liberr.Wrap(err)
			return err
		}

		redirectURI = authReq.RedirectURI
		state = authReq.State
		requestID = req.GetID()
	}

	// Generate authorization code
	code := f.generateState()
	err = f.handler.storage.SaveAuthCode(f.ctx.Request.Context(), requestID, code)
	if err != nil {
		err = liberr.Wrap(err)
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

// stringFromClaim converts a claim value to space-separated string.
func (f *IdpLogin) stringFromClaim(claimValue any) (result string) {
	var values []string

	switch v := claimValue.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}
	case []string:
		values = v
	case string:
		// Already space-separated or single value
		return v
	default:
		Log.Info("Unexpected claim type", "type", fmt.Sprintf("%T", claimValue))
		return
	}

	return strings.Join(values, " ")
}

// generateState generates a random state string.
func (f *IdpLogin) generateState() (state string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	state = base64.URLEncoding.EncodeToString(b)
	return
}
