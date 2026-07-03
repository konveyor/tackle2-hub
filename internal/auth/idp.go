package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	roleScopeRegex = regexp.MustCompile(`^\+role\.([^.\s]+)$`)
)

//
// FedIdpHandler
//

// FedIdpHandler handles external IdP federation (hub as relying party).
type FedIdpHandler struct {
	mutex    sync.Mutex
	rpClient rp.RelyingParty
	db       *gorm.DB
	storage  *Storage
	cache    *Cache
}

// Login initiates the external IdP authentication flow.
func (h *FedIdpHandler) Login(ctx *gin.Context) {
	if !federated.Idp.Enabled {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	_, err := h.RpClient()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	login := &FedIdpLogin{handler: h, ctx: ctx}
	login.begin()
}

// LoginFinished handles the redirect back from the external IdP.
func (h *FedIdpHandler) LoginFinished(ctx *gin.Context) {
	if !federated.Idp.Enabled {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	_, err := h.RpClient()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	login := &FedIdpLogin{handler: h, ctx: ctx}
	login.complete()
}

// RpClient returns the RelayParty.
func (h *FedIdpHandler) RpClient() (rpClient rp.RelyingParty, err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.rpClient != nil {
		rpClient = h.rpClient
		return
	}
	relay, err := rp.NewRelyingPartyOIDC(
		context.Background(),
		federated.Idp.Issuer,
		federated.Idp.ClientId,
		federated.Idp.ClientSecret,
		federated.Idp.RedirectURI,
		federated.Idp.Scopes,
		rp.WithHTTPClient(
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: federated.Idp.TLS,
				},
			}),
	)
	if err != nil {
		Log.Info(
			"Relay party (connect) failed.",
			"reason",
			err.Error())
	} else {
		h.rpClient = relay
		rpClient = relay
	}
	return
}

// parseAccessToken parses the access token JWT and returns the claims.
func (h *FedIdpHandler) parseAccessToken(accessToken string) (claims map[string]any, err error) {
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
func (h *FedIdpHandler) identitySaved(m *Identity) {
	h.cache.IdentitySaved(m)
}

// EndSessionURL builds the upstream IdP's end_session URL.
// The postLogoutRedirectURI is passed through to the IdP so the user
// is redirected back to the original destination after the IdP session ends.
// Returns an empty logoutURL when the IdP is not enabled or has no end_session_endpoint.
func (h *FedIdpHandler) EndSessionURL(postLogoutRedirectURI string) (logoutURL string, err error) {
	if !federated.Idp.Enabled {
		return
	}
	rpClient, err := h.RpClient()
	if err != nil {
		return
	}
	endpoint := rpClient.GetEndSessionEndpoint()
	if endpoint == "" {
		return
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	q := parsed.Query()
	q.Set("client_id", federated.Idp.ClientId)
	if postLogoutRedirectURI != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	parsed.RawQuery = q.Encode()
	logoutURL = parsed.String()
	return
}

//
// FedIdpLogin
//

// FedIdpLogin represents the state of an OIDC login flow with an external IdP.
type FedIdpLogin struct {
	handler *FedIdpHandler
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
func (f *FedIdpLogin) begin() {
	// Get authRequestID from query parameter (if present)
	f.authRequestID = f.ctx.Query(AuthRequestId)

	// Generate state and code verifier for PKCE
	f.state = f.genString()
	f.codeVerifier = f.genString()

	// Build authorization URL with PKCE
	rpClient, err := f.handler.RpClient()
	if err != nil {
		return
	}
	challenge := oidc.NewSHACodeChallenge(f.codeVerifier)
	authURL := rp.AuthURL(
		f.state,
		rpClient,
		rp.WithCodeChallenge(challenge),
	)

	// Store state and code verifier in cookies for validation in callback
	f.ctx.SetSameSite(http.SameSiteLaxMode)
	f.ctx.SetCookie("idp_state", f.state, 600, "/", "", false, true)
	f.ctx.SetCookie("idp_verifier", f.codeVerifier, 600, "/", "", false, true)

	// Store authRequestID if present (for completing OIDC flow)
	if f.authRequestID != "" {
		f.ctx.SetCookie("idp_auth_request", f.authRequestID, 600, "/", "", false, true)
	}

	// Redirect to external IdP
	f.ctx.Redirect(http.StatusFound, authURL)
}

// complete handles the redirect back from the external IdP.
func (f *FedIdpLogin) complete() {
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
func (f *FedIdpLogin) validate() (err error) {
	// Validate state (CSRF protection)
	state := f.ctx.Query("state")
	stateCookie, err := f.ctx.Cookie("idp_state")
	if err != nil || state != stateCookie {
		Log.Error(err, "State validation failed",
			"cookieErr", err != nil,
			"stateMatch", state == stateCookie)
		err = &BadRequestError{Reason: "invalid state parameter"}
		return
	}
	f.state = state

	// Get code verifier from cookie
	verifierCookie, err := f.ctx.Cookie("idp_verifier")
	if err != nil {
		Log.Error(err, "Code verifier cookie missing")
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
func (f *FedIdpLogin) exchangeCode() (err error) {
	rpClient, err := f.handler.RpClient()
	if err != nil {
		return
	}
	f.tokens, err = rp.CodeExchange[*oidc.IDTokenClaims](
		f.ctx.Request.Context(),
		f.code,
		rpClient,
		rp.WithCodeVerifier(f.codeVerifier),
	)
	if err != nil {
		Log.Error(err, "Code exchange failed")
		err = liberr.Wrap(err)
		return
	}
	return
}

// fetchUserInfo fetches user info from the IdP.
func (f *FedIdpLogin) fetchUserInfo(ctx context.Context) (err error) {
	rpClient, err := f.handler.RpClient()
	if err != nil {
		return
	}
	f.userInfo, err = rp.Userinfo[*oidc.UserInfo](
		ctx,
		f.tokens.AccessToken,
		f.tokens.TokenType,
		f.tokens.IDTokenClaims.GetSubject(),
		rpClient,
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// parseAccessToken parses the access token to extract claims.
func (f *FedIdpLogin) parseAccessToken() (err error) {
	f.accessTokenClaims, err = f.handler.parseAccessToken(f.tokens.AccessToken)
	return
}

// ensureIdentity ensures the identity created/updated.
func (f *FedIdpLogin) ensureIdentity() (err error) {
	f.identity = f.buildIdentity()
	db := f.handler.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "subject"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"kind",
			"login",
			"name",
			"email",
			"scopes",
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
func (f *FedIdpLogin) buildIdentity() (identity *Identity) {
	identity = &Identity{
		Kind:    IdentityKindOpenid,
		Issuer:  federated.Idp.Name,
		Subject: f.userInfo.Subject,
		Login:   f.userInfo.PreferredUsername,
		Email:   f.userInfo.Email,
		Name:    f.userInfo.Name,
		Scopes:  f.extractScopes(),
	}

	return
}

// extractScopes extracts scopes from access token claims.
func (f *FedIdpLogin) extractScopes() (scopes []string) {
	if f.accessTokenClaims == nil {
		return
	}
	if scopeClaim, found := f.accessTokenClaims[ClaimScope]; found {
		str := f.asString(scopeClaim)
		scopes = strings.Fields(str)
		scopes = f.expandScopes(scopes)
		scopes = ExpandScopes(scopes...)
	}
	return
}

// expandScopes expands role references in the scope list.
// Role references use +role.name syntax and are replaced with the role's
// permission scopes. Reduces the burden of scope management in IdPs.
func (f *FedIdpLogin) expandScopes(in []string) (expanded []string) {
	cache := f.handler.cache
	for _, scope := range in {
		match := roleScopeRegex.FindStringSubmatch(scope)
		if len(match) < 2 {
			expanded = append(expanded, scope)
			continue
		}
		name := match[1]
		role, err := cache.FindRoleByName(name)
		if err != nil {
			Log.Info(err.Error(),
				"pattern",
				scope)
			continue
		}
		expanded = append(expanded, role.Scopes...)
	}
	return
}

// issueTokens creates a hub OIDC session and redirects to UI.
func (f *FedIdpLogin) issueTokens() (err error) {
	var redirectURI string
	var state string
	var requestID string

	// Use external IdP subject directly
	subject := f.identity.Subject

	if f.authRequestID == "" {
		err = fmt.Errorf("missing authorization request ID")
		return
	}

	// Complete existing OIDC authorization request
	var authReq op.AuthRequest
	authReq, err = f.handler.storage.AuthRequestByID(f.ctx.Request.Context(), f.authRequestID)
	if err != nil {
		return
	}

	// Update the auth request with the authenticated user
	if ar, cast := authReq.(*AuthRequest); cast {
		ar.subject = subject
		ar.idpRefreshToken = f.tokens.RefreshToken
		ar.idpIdentityId = f.identity.ID
	}

	redirectURI = authReq.GetRedirectURI()
	state = authReq.GetState()
	requestID = f.authRequestID

	// Generate authorization code
	code := f.genString()
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
func (f *FedIdpLogin) asString(claim any) (s string) {
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

// refreshIdentity refreshes an IdpIdentity using its refresh token.
func (f *FedIdpLogin) refreshIdentity(grant *Grant) (err error) {
	rpClient, err := f.handler.RpClient()
	if err != nil {
		return
	}
	f.tokens, err = rp.RefreshTokens[*oidc.IDTokenClaims](
		context.Background(),
		rpClient,
		grant.IdpRefreshToken,
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
	grant.IdpRefreshToken = f.tokens.RefreshToken
	err = f.ensureIdentity()
	return
}

// genString generates a random string.
func (f *FedIdpLogin) genString() (s string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	s = base64.RawURLEncoding.EncodeToString(b)
	return
}
