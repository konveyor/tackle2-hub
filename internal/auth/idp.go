package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"gorm.io/gorm"
)

// IdpHandler handles external IdP federation (hub as relying party).
type IdpHandler struct {
	rpClient rp.RelyingParty
	db       *gorm.DB
	storage  *Storage
}

// ServeHTTP routes requests to appropriate handlers.
func (h *IdpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !Settings.Auth.Idp.Enabled {
		http.Error(w, "external IdP not configured", http.StatusNotFound)
		return
	}

	switch r.URL.Path {
	case "/login":
		h.login(w, r)
	case "/callback":
		h.callback(w, r)
	default:
		http.NotFound(w, r)
	}
}

// login initiates the external IdP authentication flow.
func (h *IdpHandler) login(w http.ResponseWriter, r *http.Request) {
	// Get authRequestID from query parameter (if present)
	authRequestID := r.URL.Query().Get("authRequestID")

	// Generate state and code verifier for PKCE
	state := h.generateState()
	codeVerifier := h.generateState()

	// Build authorization URL with PKCE
	authURL := rp.AuthURL(
		state,
		h.rpClient,
		rp.WithCodeChallenge(oidc.NewSHACodeChallenge(codeVerifier)),
	)

	// Store state and code verifier in cookies for validation in callback
	http.SetCookie(w, &http.Cookie{
		Name:     "idp_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "idp_verifier",
		Value:    codeVerifier,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Store authRequestID if present (for completing OIDC flow)
	if authRequestID != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "idp_auth_request",
			Value:    authRequestID,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	// Redirect to external IdP
	http.Redirect(w, r, authURL, http.StatusFound)
}

// callback handles the redirect back from the external IdP.
func (h *IdpHandler) callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate state (CSRF protection)
	state := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("idp_state")
	if err != nil || state != stateCookie.Value {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	// Get code verifier from cookie
	verifierCookie, err := r.Cookie("idp_verifier")
	if err != nil {
		http.Error(w, "missing code verifier", http.StatusBadRequest)
		return
	}

	// Get authRequestID from cookie (if present)
	var authRequestID string
	if authReqCookie, err := r.Cookie("idp_auth_request"); err == nil {
		authRequestID = authReqCookie.Value
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:   "idp_state",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "idp_verifier",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "idp_auth_request",
		Path:   "/",
		MaxAge: -1,
	})

	// Check for error from IdP
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf("IdP error: %s - %s", errParam, errDesc), http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](
		ctx,
		code,
		h.rpClient,
		rp.WithCodeVerifier(verifierCookie.Value),
	)
	if err != nil {
		Log.Error(err, "code exchange failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Get user info from IdP
	userInfo, err := rp.Userinfo[*oidc.UserInfo](
		ctx,
		tokens.AccessToken,
		tokens.TokenType,
		tokens.IDTokenClaims.GetSubject(),
		h.rpClient,
	)
	if err != nil {
		Log.Error(err, "userinfo request failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Parse access token to extract claims
	accessTokenClaims, err := h.parseAccessToken(tokens.AccessToken)
	if err != nil {
		Log.Error(err, "access token parsing failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Find or create IdP identity
	idpIdentity, err := h.findOrCreateIdpIdentity(ctx, userInfo, tokens, accessTokenClaims)
	if err != nil {
		Log.Error(err, "IdP identity creation/update failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	Log.Info("External IdP authentication successful", "subject", idpIdentity.Subject)

	// Issue hub tokens (complete OIDC flow)
	err = h.issueHubTokens(ctx, w, r, idpIdentity, authRequestID)
	if err != nil {
		Log.Error(err, "token issuance failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}
}

// findOrCreateIdpIdentity finds existing identity or creates new one from IdP userinfo.
func (h *IdpHandler) findOrCreateIdpIdentity(
	ctx context.Context,
	userInfo *oidc.UserInfo,
	tokens *oidc.Tokens[*oidc.IDTokenClaims],
	accessTokenClaims map[string]any) (idpIdentity *model.IdpIdentity, err error) {
	//
	idpIdentity = &model.IdpIdentity{}
	result := h.db.First(
		idpIdentity,
		"Provider = ? AND Subject = ?",
		Settings.Auth.Idp.Name,
		userInfo.Subject,
	)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// First time login - create identity
		idpIdentity, err = h.createIdpIdentity(ctx, userInfo, tokens, accessTokenClaims)
		if err != nil {
			return
		}
	} else if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	} else {
		// Returning user - update identity
		err = h.updateIdpIdentity(ctx, idpIdentity, tokens, userInfo, accessTokenClaims)
		if err != nil {
			return
		}
	}

	return
}

// createIdpIdentity creates a new IdpIdentity for first-time login.
func (h *IdpHandler) createIdpIdentity(
	ctx context.Context,
	userInfo *oidc.UserInfo,
	tokens *oidc.Tokens[*oidc.IDTokenClaims],
	accessTokenClaims map[string]any) (idpIdentity *model.IdpIdentity, err error) {
	//
	// Extract user details from userinfo
	email := userInfo.Email
	if email == "" {
		email = fmt.Sprintf("%s@external", userInfo.Subject)
	}

	userid := userInfo.PreferredUsername
	if userid == "" {
		userid = userInfo.Subject
	}

	// Extract scopes and roles from access token
	scopes, roles := h.extractClaimsFromAccessToken(accessTokenClaims)

	// Create IdpIdentity
	expiration := time.Now().Add(
		time.Duration(Settings.Auth.Token.RefreshLifespan) * time.Second,
	)
	if tokens.Expiry.After(time.Now()) {
		expiration = tokens.Expiry
	}

	idpIdentity = &model.IdpIdentity{
		Provider:          Settings.Auth.Idp.Name,
		Subject:           userInfo.Subject,
		Userid:            userid,
		Email:             email,
		RefreshToken:      tokens.RefreshToken,
		Expiration:        expiration,
		LastAuthenticated: time.Now(),
		LastRefreshed:     time.Now(),
		Scopes:            scopes,
		Roles:             roles,
	}
	idpIdentity.CreateUser = "system"

	err = h.db.Create(idpIdentity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// updateIdpIdentity updates an existing IdpIdentity with new tokens and userinfo.
func (h *IdpHandler) updateIdpIdentity(
	ctx context.Context,
	idpIdentity *model.IdpIdentity,
	tokens *oidc.Tokens[*oidc.IDTokenClaims],
	userInfo *oidc.UserInfo,
	accessTokenClaims map[string]any) (err error) {
	//
	// Update user details from userinfo
	if userInfo.Email != "" {
		idpIdentity.Email = userInfo.Email
	}
	if userInfo.PreferredUsername != "" {
		idpIdentity.Userid = userInfo.PreferredUsername
	}

	// Extract scopes and roles from access token
	scopes, roles := h.extractClaimsFromAccessToken(accessTokenClaims)

	// Update tokens, timestamps, and claims
	idpIdentity.RefreshToken = tokens.RefreshToken
	if tokens.Expiry.After(time.Now()) {
		idpIdentity.Expiration = tokens.Expiry
	}
	idpIdentity.LastAuthenticated = time.Now()
	idpIdentity.LastRefreshed = time.Now()
	idpIdentity.Scopes = scopes
	idpIdentity.Roles = roles
	idpIdentity.UpdateUser = "system"

	err = h.db.Save(idpIdentity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// extractClaimsFromAccessToken extracts scopes and roles from access token claims.
func (h *IdpHandler) extractClaimsFromAccessToken(
	claims map[string]any) (scopes string, roles string) {
	//
	if claims == nil {
		return
	}

	// Extract scopes
	if scopeClaim, found := claims["scope"]; found {
		scopes = h.stringFromClaim(scopeClaim)
	}

	// Extract roles
	if roleClaim, found := claims["roles"]; found {
		roles = h.stringFromClaim(roleClaim)
	}

	return
}

// stringFromClaim converts a claim value to space-separated string.
func (h *IdpHandler) stringFromClaim(claimValue any) (result string) {
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

// issueHubTokens creates a hub OIDC session and redirects to UI.
func (h *IdpHandler) issueHubTokens(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	idpIdentity *model.IdpIdentity,
	authRequestID string) (err error) {
	//
	var redirectURI string
	var state string
	var requestID string

	// Build subject from IdP identity
	subject := h.idpSubject(idpIdentity.Subject)

	if authRequestID != "" {
		// Complete existing OIDC authorization request
		authReq, err := h.storage.AuthRequestByID(ctx, authRequestID)
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
		requestID = authRequestID
	} else {
		// Create a new authorization request (standalone IdP login)
		authReq := &oidc.AuthRequest{
			ClientID:     Settings.Auth.Client.ID,
			RedirectURI:  Settings.Auth.Client.RedirectURIs[0],
			Scopes:       []string{"openid", "profile", "email"},
			ResponseType: oidc.ResponseTypeCode,
			State:        h.generateState(),
		}

		req, err := h.storage.CreateAuthRequest(ctx, authReq, subject)
		if err != nil {
			err = liberr.Wrap(err)
			return err
		}

		redirectURI = authReq.RedirectURI
		state = authReq.State
		requestID = req.GetID()
	}

	// Generate authorization code
	code := h.generateState()
	err = h.storage.SaveAuthCode(ctx, requestID, code)
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
	http.Redirect(w, r, redirectURL, http.StatusFound)
	return
}

// idpSubject builds the hub subject from IdP subject.
func (h *IdpHandler) idpSubject(subject string) (s string) {
	s = fmt.Sprintf("idp:%s:%s", Settings.Auth.Idp.Name, subject)
	return
}

// generateState generates a random state string.
func (h *IdpHandler) generateState() (state string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	state = base64.URLEncoding.EncodeToString(b)
	return
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
