package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

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

	// Find or create user
	user, err := h.findOrCreateUser(ctx, userInfo, tokens)
	if err != nil {
		Log.Error(err, "user creation/update failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Issue hub tokens
	err = h.issueHubTokens(ctx, w, r, user)
	if err != nil {
		Log.Error(err, "token issuance failed")
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}
}

// findOrCreateUser finds existing user or creates new one from IdP userinfo.
func (h *IdpHandler) findOrCreateUser(
	ctx context.Context,
	userInfo *oidc.UserInfo,
	tokens *oidc.Tokens[*oidc.IDTokenClaims]) (user *model.User, err error) {
	//
	var idpIdentity model.IdpIdentity
	result := h.db.First(
		&idpIdentity,
		"provider = ? AND subject = ?",
		Settings.Auth.Idp.Name,
		userInfo.Subject,
	)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// First time login - create user and identity
		user, err = h.createUser(ctx, userInfo, tokens)
		if err != nil {
			return
		}
	} else if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	} else {
		// Returning user - update identity
		err = h.updateIdpIdentity(ctx, &idpIdentity, tokens)
		if err != nil {
			return
		}

		// Load the user
		user = &model.User{}
		err = h.db.First(user, idpIdentity.UserID).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	return
}

// createUser creates a new user and IdpIdentity for first-time login.
func (h *IdpHandler) createUser(
	ctx context.Context,
	userInfo *oidc.UserInfo,
	tokens *oidc.Tokens[*oidc.IDTokenClaims]) (user *model.User, err error) {
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

	// Create user
	user = &model.User{
		Subject:  h.idpSubject(userInfo.Subject),
		Userid:   userid,
		Email:    email,
		Password: h.generateState(), // Random password (won't be used)
	}
	user.CreateUser = "system"

	err = h.db.Create(user).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	// Create IdpIdentity
	expiration := time.Now().Add(
		time.Duration(Settings.Auth.Token.RefreshLifespan) * time.Second,
	)
	if tokens.Expiry.After(time.Now()) {
		expiration = tokens.Expiry
	}

	idpIdentity := &model.IdpIdentity{
		Provider:          Settings.Auth.Idp.Name,
		Subject:           userInfo.Subject,
		RefreshToken:      tokens.RefreshToken,
		Expiration:        expiration,
		LastAuthenticated: time.Now(),
		LastRefreshed:     time.Now(),
		UserID:            user.ID,
	}
	idpIdentity.CreateUser = "system"

	err = h.db.Create(idpIdentity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// updateIdpIdentity updates an existing IdpIdentity with new tokens.
func (h *IdpHandler) updateIdpIdentity(
	ctx context.Context,
	idpIdentity *model.IdpIdentity,
	tokens *oidc.Tokens[*oidc.IDTokenClaims]) (err error) {
	//
	// Update tokens and timestamps
	idpIdentity.RefreshToken = tokens.RefreshToken
	if tokens.Expiry.After(time.Now()) {
		idpIdentity.Expiration = tokens.Expiry
	}
	idpIdentity.LastAuthenticated = time.Now()
	idpIdentity.LastRefreshed = time.Now()
	idpIdentity.UpdateUser = "system"

	err = h.db.Save(idpIdentity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// issueHubTokens creates a hub OIDC session and redirects to UI.
func (h *IdpHandler) issueHubTokens(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	user *model.User) (err error) {
	//
	// Create an authorization request in hub's OP
	authReq := &oidc.AuthRequest{
		ClientID:     Settings.Auth.Client.ID,
		RedirectURI:  Settings.Auth.Client.RedirectURIs[0],
		Scopes:       []string{"openid", "profile", "email"},
		ResponseType: oidc.ResponseTypeCode,
		State:        h.generateState(),
	}

	// Create the auth request in storage
	req, err := h.storage.CreateAuthRequest(
		ctx,
		authReq,
		user.Subject,
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	// Generate authorization code
	code := h.generateState()
	err = h.storage.SaveAuthCode(ctx, req.GetID(), code)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	// Build redirect URL with code
	redirectURL := fmt.Sprintf(
		"%s?code=%s&state=%s",
		authReq.RedirectURI,
		code,
		authReq.State,
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
