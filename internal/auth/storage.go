package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-jose/go-jose/v4"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Storage implements op.Storage.
type Storage struct {
	mutex               sync.Mutex
	keySet              KeySet
	db                  *gorm.DB
	authReqById         map[string]*AuthRequest
	authReqByCode       map[string]string
	devAuthReqByDevCode map[string]*DeviceAuthRequest
	devAuthByUserCode   map[string]string
	idpHandler          *FedIdpHandler
	dsHandler           *LdapHandler
	cache               *Cache
}

// GetClientByClientID retrieves a client by ID from database.
func (r *Storage) GetClientByClientID(ctx context.Context, clientId string) (opClient op.Client, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	req, cast := ctx.Value(ReqInCtx).(*http.Request)
	if !cast {
		err = oidc.ErrServerError().
			WithDescription("ctx must contain request")
		return
	}
	if clientId == DevVerifierClientId {
		opClient = &Client{
			id:              DevVerifierClientId,
			subject:         "device-verifier-subject",
			applicationType: op.ApplicationTypeWeb,
			grantTypes:      []string{"authorization_code"},
			redirectURIs:    []string{AppendIssuer(req, api.DeviceCbRoute)},
			scopes:          []string{"openid"},
			request:         req,
		}
		return
	}
	m, err := r.cache.FindClientByStrId(clientId)
	if err != nil {
		if errors.Is(err, &NotFound{}) {
			err = oidc.ErrInvalidClient().
				WithDescription("%s", err.Error())
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	client := &Client{}
	client.With(m, req)
	client.Inject()
	opClient = client
	return
}

// AuthorizeClientIDSecret validates client credentials.
func (r *Storage) AuthorizeClientIDSecret(ctx context.Context, id, passphrase string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	client, err := r.GetClientByClientID(ctx, id)
	if err != nil {
		return
	}
	found := client.(*Client)
	if found.secret == "" {
		return
	}
	if !secret.MatchPassword(passphrase, found.secret) {
		err = oidc.ErrInvalidClient().WithDescription("clientSecret not-valid.")
	}
	return
}

// ClientCredentials validates client credentials for client credentials flow.
func (r *Storage) ClientCredentials(ctx context.Context, id, passphrase string) (client op.Client, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	client, err = r.GetClientByClientID(ctx, id)
	if err != nil {
		return
	}
	found := client.(*Client)
	if found.secret == "" {
		return
	}
	if !secret.MatchPassword(passphrase, found.secret) {
		err = oidc.ErrInvalidClient().WithDescription("clientSecret not-valid.")
	}
	return
}

// ClientCredentialsTokenRequest creates a token request for client credentials.
func (r *Storage) ClientCredentialsTokenRequest(
	ctx context.Context,
	clientId string,
	scopes []string) (req op.TokenRequest, err error) {
	//
	client, err := r.GetClientByClientID(ctx, clientId)
	if err != nil {
		return
	}
	c := client.(*Client)
	req = &ClientRequest{
		authId:   r.genId(),
		clientId: clientId,
		subject:  c.subject,
		scopes:   scopes,
		issued:   time.Now(),
	}
	return
}

// CreateAuthRequest initiates an authorization request.
func (r *Storage) CreateAuthRequest(
	_ context.Context,
	authReq *oidc.AuthRequest,
	userID string) (req op.AuthRequest, err error) {
	//
	now := time.Now()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for id, req := range r.authReqById {
		if now.After(req.expiration) {
			delete(r.authReqById, id)
			delete(r.authReqByCode, req.authCode)
		}
	}
	requestId := r.genId()
	found := false
	for _, scope := range authReq.Scopes {
		if scope == oidc.ScopeOfflineAccess {
			found = true
			break
		}
	}
	if !found {
		authReq.Scopes = append(authReq.Scopes, oidc.ScopeOfflineAccess)
	}
	req = &AuthRequest{
		AuthRequest: authReq,
		requestId:   requestId,
		subject:     userID,
		issued:      now,
		expiration:  now.Add(time.Hour),
	}
	r.authReqById[requestId] = req.(*AuthRequest)
	return
}

// AuthRequestByID retrieves an auth request by ID.
func (r *Storage) AuthRequestByID(_ context.Context, id string) (req op.AuthRequest, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	authReq, found := r.authReqById[id]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest not-found.")
		return
	}
	if time.Now().After(authReq.expiration) {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest expired")
		return
	}
	req = authReq
	return
}

// AuthRequestByCode retrieves auth request by authorization code.
func (r *Storage) AuthRequestByCode(_ context.Context, code string) (req op.AuthRequest, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	requestId, found := r.authReqByCode[code]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("authCode not-found.")
		return
	}
	authReq, found := r.authReqById[requestId]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest not-found.")
		return
	}
	if time.Now().After(authReq.expiration) {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest expired")
		return
	}
	req = authReq
	return
}

// SaveAuthCode stores the authorization code.
func (r *Storage) SaveAuthCode(_ context.Context, id, code string) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	authReq, found := r.authReqById[id]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest not-found.")
		return
	}
	authReq.authCode = code
	r.authReqByCode[code] = id
	return
}

// DeleteAuthRequest deletes an auth request.
func (r *Storage) DeleteAuthRequest(_ context.Context, id string) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	authReq, found := r.authReqById[id]
	if !found {
		return
	}
	delete(r.authReqByCode, authReq.authCode)
	delete(r.authReqById, id)
	return
}

// CreateAccessToken creates an access token.
// For AuthRequest: Creates a grant using req.GetID() as the authId, then creates
// a token with the same authId and links it to the grant via GrantID.
// For RefreshRequest: Uses the existing grant's authId. The upsert on authId
// updates the existing token instead of creating a new one.
// For ClientRequest: Creates a token with no associated grant.
// For DeviceAuthorizationState: Creates a grant and token for device authorization flow.
func (r *Storage) CreateAccessToken(
	ctx context.Context,
	req op.TokenRequest) (tokenId string, expiration time.Time, err error) {
	//
	err = r.injectScopes(req)
	if err != nil {
		return
	}
	subject := req.GetSubject()
	s, err := r.findSubject(subject)
	if err != nil {
		err = oidc.ErrInvalidGrant().
			WithDescription("%s", err.Error())
		return
	}
	var authId string
	var grantId string
	expiration = time.Now().Add(Settings.Token.Lifespan)
	switch req := req.(type) {
	case *AuthRequest:
		authId = req.GetID()
		grantId = authId
		authCode := r.authCodeById(authId)
		_, err = r.createGrant(ctx, req, authId, authCode)
		if err != nil {
			return
		}
	case *RefreshRequest:
		authId = req.grantId
		grantId = authId
	case *ClientRequest:
		authId = req.authId
	case *op.DeviceAuthorizationState:
		authId = r.genId()
		grantId = authId
		_, err = r.createGrant(ctx, req, authId, "")
		if err != nil {
			return
		}
	default:
		err = oidc.ErrServerError().
			WithDescription("unsupported token request type")
		return
	}
	tokenId = authId
	m := &Token{}
	m.Kind = KindAccessToken
	m.AuthId = authId
	m.Subject = subject
	m.Scopes = req.GetScopes()
	m.Issued = time.Now()
	m.Expiration = expiration
	m.GrantID = r.grantId(grantId)
	if s != nil {
		m.UserID = s.UserId
		m.IdpIdentityID = s.IdentityId
		m.IdpClientID = s.ClientId
	}
	db := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "authId"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"expiration",
			"scopes",
		}),
	})
	err = db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	if len(s.Scopes) == 0 {
		Log.Info(
			"WARNING: issued (access) token has no scopes.",
			"login", s.Login(),
			"authId", m.AuthId,
			"id", m.ID)
	}
	return
}

// CreateAccessAndRefreshTokens creates both access and refresh tokens.
func (r *Storage) CreateAccessAndRefreshTokens(
	ctx context.Context,
	req op.TokenRequest,
	currentRefresh string) (
	accessTokenId string,
	refreshToken string,
	expiration time.Time,
	err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	err = r.injectScopes(req)
	if err != nil {
		return
	}
	accessTokenId, expiration, err = r.CreateAccessToken(ctx, req)
	if err != nil {
		return
	}
	refreshToken, err = r.createRefreshToken(ctx, req)
	if err != nil {
		return
	}
	if refreshToken == "" {
		refreshToken = currentRefresh
	}
	return
}

// TokenRequestByRefreshToken retrieves token request by refresh token.
func (r *Storage) TokenRequestByRefreshToken(
	ctx context.Context,
	refreshToken string) (req op.RefreshTokenRequest, err error) {
	//
	defer func() {
		if err != nil {
			if !errors.Is(err, oidc.ErrInvalidGrant()) {
				Log.Error(err, "")
			}
		}
	}()
	grant, err := r.grantByRefreshToken(ctx, refreshToken)
	if err != nil {
		return
	}
	req = &RefreshRequest{
		grantId:  grant.AuthId,
		clientId: grant.ClientId,
		subject:  grant.Subject,
		scopes:   grant.Scopes,
		issued:   grant.Issued,
	}
	err = r.refreshIdentity(req, grant)
	if err == nil {
		return
	}
	if errors.Is(err, oidc.ErrInvalidGrant()) {
		Log.Error(
			r.deleteGrant(ctx, grant.AuthId),
			"Grant delete failed.")
	}
	return
}

// TerminateSession terminates a user session.
func (r *Storage) TerminateSession(ctx context.Context, userID, _ string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	var grants []Grant
	err = r.db.Find(&grants, "subject", userID).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range grants {
		grant := &grants[i]
		err = r.deleteGrant(ctx, grant.AuthId)
		if err != nil {
			return
		}
	}
	err = r.deleteTokensBySubject(ctx, userID)
	if err != nil {
		return
	}
	return
}

// TerminateSessionFromRequest terminates a user session and returns the redirect URL.
// For federated (OIDC) users, the redirect chains through the upstream IdP's
// end_session_endpoint so that both the hub and IdP sessions are terminated.
func (r *Storage) TerminateSessionFromRequest(
	ctx context.Context,
	req *op.EndSessionRequest) (redirect string, err error) {
	//
	err = r.TerminateSession(ctx, req.UserID, req.ClientID)
	if err != nil {
		return
	}
	redirect = req.RedirectURI
	if req.UserID == "" {
		return
	}
	s, err := r.findSubject(req.UserID)
	if err != nil {
		if errors.Is(err, &NotFound{}) {
			err = nil
		}
		return
	}
	if !s.IsIdentity() || s.Identity.Kind != IdentityKindOpenid {
		return
	}
	logoutURL, err := r.idpHandler.EndSessionURL(req.RedirectURI)
	if err != nil {
		return
	}
	if logoutURL != "" {
		redirect = logoutURL
	}
	return
}

// RevokeToken revokes a token.
func (r *Storage) RevokeToken(
	ctx context.Context,
	tokenRef string,
	userId string,
	clientId string) (errPtr *oidc.Error) {
	//
	defer func() {
		if errPtr != nil {
			Log.Error(errPtr, "")
		}
	}()
	digest := secret.Hash(tokenRef)
	grant := &Grant{}
	err := r.db.First(grant, "refreshToken", digest).Error
	if err == nil {
		err = r.deleteGrant(ctx, grant.AuthId)
		if err != nil {
			errPtr = oidc.ErrServerError()
			return
		}
		return
	}
	err = r.deleteToken(ctx, tokenRef)
	if err != nil {
		errPtr = oidc.ErrServerError()
		return
	}
	return
}

// SigningKey returns the current signing key.
func (r *Storage) SigningKey(_ context.Context) (key op.SigningKey, err error) {
	key = r.keySet.SigningKey()
	return
}

// SignatureAlgorithms returns supported signature algorithms.
func (r *Storage) SignatureAlgorithms(ctx context.Context) (alg []jose.SignatureAlgorithm, err error) {
	alg = []jose.SignatureAlgorithm{jose.RS256}
	return
}

// KeySet returns the JWKS.
func (r *Storage) KeySet(ctx context.Context) (keys []op.Key, err error) {
	keys = make([]op.Key, 0)
	for _, jwk := range r.keySet.Keys {
		keys = append(
			keys,
			&Key{
				jwk: jwk,
			})
	}
	return
}

// GetKeyByIDAndClientID retrieves a key.
func (r *Storage) GetKeyByIDAndClientID(_ context.Context, keyID, clientId string) (key *jose.JSONWebKey, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	jwk, err := r.keySet.Key(keyID)
	if err != nil {
		return
	}
	key = &jose.JSONWebKey{
		KeyID:     jwk.KeyID,
		Algorithm: jwk.Algorithm,
		Use:       jwk.Use,
		Key:       jwk.PrivateKey,
	}
	return
}

// ValidateJWTProfileScopes validates JWT profile scopes.
func (r *Storage) ValidateJWTProfileScopes(_ context.Context, _ string, scopes []string) (valid []string, err error) {
	valid = scopes
	return
}

// Health checks storage health.
func (r *Storage) Health(_ context.Context) (err error) {
	err = r.db.Exec("SELECT 1").Error
	return
}

// GetPrivateClaimsFromScopes returns private claims based on scopes.
func (r *Storage) GetPrivateClaimsFromScopes(
	_ context.Context,
	userID string,
	clientId string,
	scopes []string) (claims map[string]any, err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	claims = make(map[string]any)
	if userID == "" {
		return
	}

	s, err := r.findSubject(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	if s == nil {
		return
	}

	// Add scopes to JWT claims
	// The scopes parameter already contains both standard OAuth scopes (openid, offline_access)
	// and our injected permission scopes from injectScopes()
	if len(scopes) > 0 {
		claims[ClaimScope] = strings.Join(scopes, " ")
	}

	return
}

// SetUserinfoFromScopes is deprecated but required by the interface.
func (r *Storage) SetUserinfoFromScopes(
	_ context.Context,
	userinfo *oidc.UserInfo,
	userId string,
	clientId string,
	scopes []string) (err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()

	s, err := r.findSubject(userId)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	userinfo.Subject = userId
	// Set PreferredUsername from the appropriate source
	if s.User != nil {
		userinfo.PreferredUsername = s.User.Login
	} else if s.Identity != nil {
		userinfo.PreferredUsername = s.Identity.Login
	} else {
		userinfo.PreferredUsername = s.Key
	}
	if s.Email != "" {
		userinfo.Email = s.Email
		userinfo.EmailVerified = true
	}
	return
}

// SetUserinfoFromToken sets userinfo claims.
func (r *Storage) SetUserinfoFromToken(
	_ context.Context,
	userinfo *oidc.UserInfo,
	tokenId string,
	subject string,
	origin string) (err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()

	s, err := r.findSubject(subject)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if s == nil {
		return
	}

	userinfo.Subject = subject
	// Set PreferredUsername from the appropriate source
	if s.User != nil {
		userinfo.PreferredUsername = s.User.Login
	} else if s.Identity != nil {
		userinfo.PreferredUsername = s.Identity.Login
	} else {
		userinfo.PreferredUsername = s.Key
	}
	if s.Email != "" {
		userinfo.Email = s.Email
		userinfo.EmailVerified = true
	}
	return
}

// SetIntrospectionFromToken sets introspection response.
func (r *Storage) SetIntrospectionFromToken(
	ctx context.Context,
	introspection *oidc.IntrospectionResponse,
	tokenId string,
	subject string,
	clientId string) (err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	token, err := r.token(ctx, tokenId)
	if err != nil {
		return
	}
	expiration := int(token.Expiration.Unix())
	introspection.Active = expiration > int(time.Now().Unix())
	introspection.Scope = token.Scopes
	introspection.ClientID = clientId
	introspection.Subject = token.Subject
	introspection.Expiration = oidc.FromTime(token.Expiration)
	return
}

// GetRefreshTokenInfo retrieves refresh token info.
func (r *Storage) GetRefreshTokenInfo(
	ctx context.Context,
	clientId string,
	token string) (userId, tokenId string, err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	grant, err := r.grantByRefreshToken(ctx, token)
	if err != nil {
		return
	}
	userId = grant.Subject
	tokenId = token
	return
}

// Login handles the authentication flow.
func (r *Storage) Login(writer http.ResponseWriter, request *http.Request, authReqId string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	if federated.Idp.Enabled && federated.Idp.Primary {
		loginURL := AppendIssuer(request, api.IdpLoginRoute)
		parsedURL, _ := url.Parse(loginURL)
		query := parsedURL.Query()
		query.Set(AuthRequestId, authReqId)
		parsedURL.RawQuery = query.Encode()
		http.Redirect(writer, request, parsedURL.String(), http.StatusFound)
		return
	}
	login := &Login{
		storage:   r,
		writer:    writer,
		request:   request,
		authReqId: authReqId,
	}
	err = login.complete()
	return
}

//
// Login
//

// Login represents the state of a user login flow.
type Login struct {
	storage   *Storage
	writer    http.ResponseWriter
	request   *http.Request
	authReqId string

	login    string
	password string
	subject  *Subject
	authReq  *AuthRequest
}

// complete handles the login form submission and authentication.
func (r *Login) complete() (err error) {
	err = r.parseCredentials()
	if err != nil {
		return
	}

	if r.login == "" || r.password == "" {
		err = r.renderPage()
		return
	}

	err = r.authenticate()
	if err != nil {
		Log.Info(err.Error())
		_ = r.renderPage()
		err = nil
		return
	}

	err = r.updateAuthRequest()
	if err != nil {
		return
	}

	r.redirect()
	return
}

// parseCredentials extracts credentials from the form.
func (r *Login) parseCredentials() (err error) {
	err = r.request.ParseForm()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.login = r.request.PostFormValue("login")
	r.password = r.request.PostFormValue("password")
	return
}

// authenticate validates the user credentials.
// Sets the Login.subject.
func (r *Login) authenticate() (err error) {
	for _, method := range []func() error{
		r.authUser,
		r.authLdapUser,
	} {
		err = method()
		if err == nil {
			// authenticated
			return
		}
		if errors.Is(err, &NotFound{}) {
			// next
			continue
		}
		if errors.Is(err, &NotAuthenticated{}) {
			// rejected
			break
		}
	}
	return
}

// authUser authenticates a user.
func (r *Login) authUser() (err error) {
	cache := r.storage.cache
	user, err := cache.FindUserByLogin(r.login)
	if err == nil {
		if !secret.MatchPassword(r.password, user.Password) {
			err = &NotAuthenticated{
				Reason: "invalid password",
				Token:  r.login,
			}
			return
		}
		var scopes []string
		scopes, err = cache.FindScopes(user.Subject)
		if err != nil {
			return
		}
		subject := &Subject{}
		subject.WithUser(user, scopes)
		r.subject = subject
		return
	}
	return
}

// authUser authenticates an LDAP user.
func (r *Login) authLdapUser() (err error) {
	r.subject, err = r.storage.dsHandler.Authenticate(r.login, r.password)
	return
}

// updateAuthRequest updates the auth request with authenticated user.
func (r *Login) updateAuthRequest() (err error) {
	r.storage.mutex.Lock()
	defer r.storage.mutex.Unlock()

	var found bool
	r.authReq, found = r.storage.authReqById[r.authReqId]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest not-found.")
		return
	}
	if time.Now().After(r.authReq.expiration) {
		err = oidc.ErrInvalidGrant().WithDescription("authRequest expired")
		return
	}

	r.authReq.subject = r.subject.Key
	r.authReq.idpRefreshToken = r.password
	r.authReq.issued = time.Now()
	r.authReq.done = true
	return
}

// redirect redirects to the authorization callback.
func (r *Login) redirect() {
	issuer := AppendIssuer(r.request, api.AuthorizeCbRoute)
	cbURL := fmt.Sprintf("%s?id=%s", issuer, r.authReqId)
	http.Redirect(r.writer, r.request, cbURL, http.StatusFound)
}

// renderPage renders the login page.
func (r *Login) renderPage() (err error) {
	idpButton := ""
	if federated.Idp.Enabled {
		loginURL := AppendIssuer(r.request, api.IdpLoginRoute)
		parsedURL, pErr := url.Parse(loginURL)
		if pErr != nil {
			err = liberr.Wrap(pErr)
			return
		}
		query := parsedURL.Query()
		query.Set(AuthRequestId, r.authReqId)
		parsedURL.RawQuery = query.Encode()

		idpButton = `
        <div class="divider">OR</div>
        <a href="` + parsedURL.String() + `" class="idp-button">
            Sign in with ` + federated.Idp.Name + `
        </a>`
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Tackle Hub - Login</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            padding: 40px;
            max-width: 400px;
            width: 100%;
        }
        h1 {
            color: #333;
            font-size: 20px;
            margin-bottom: 8px;
        }
        .subtitle {
            color: #666;
            font-size: 13px;
            margin-bottom: 30px;
        }
        label {
            display: block;
            color: #333;
            font-size: 13px;
            font-weight: 500;
            margin-bottom: 8px;
        }
        input[type="text"],
        input[type="password"] {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.2s;
            margin-bottom: 16px;
        }
        input[type="text"]:focus,
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 12px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 15px;
            font-weight: 500;
            cursor: pointer;
            transition: transform 0.1s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
        .divider {
            margin: 24px 0;
            text-align: center;
            color: #999;
            font-size: 13px;
            position: relative;
        }
        .divider::before,
        .divider::after {
            content: '';
            position: absolute;
            top: 50%;
            width: 40%;
            height: 1px;
            background: #e0e0e0;
        }
        .divider::before { left: 0; }
        .divider::after { right: 0; }
        .idp-button {
            display: block;
            width: 100%;
            padding: 12px;
            background: #28a745;
            color: white;
            text-decoration: none;
            border-radius: 8px;
            text-align: center;
            font-size: 15px;
            font-weight: 500;
            transition: transform 0.1s, box-shadow 0.2s;
        }
        .idp-button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(40, 167, 69, 0.4);
        }
        .idp-button:active {
            transform: translateY(0);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Tackle Login</h1>
        <p class="subtitle">Sign in to continue</p>
        <form action="` + AppendIssuer(r.request, api.LoginRoute) + `?` + AuthRequestId + `=` + r.authReqId + `" method="post">
            <label for="login">Login</label>
            <input type="text" id="login" name="login" required autofocus>

            <label for="password">Password</label>
            <input type="password" id="password" name="password" required>

            <button type="submit">Sign In</button>
        </form>` + idpButton + `
    </div>
</body>
</html>`
	r.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = r.writer.Write([]byte(html))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// injectScopes adds user/identity permissions as scopes to the token request.
func (r *Storage) injectScopes(req op.TokenRequest) (err error) {
	subject := req.GetSubject()
	if subject == "" {
		return
	}
	s, err := r.findSubject(subject)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	if s == nil {
		return
	}
	scopes := append(req.GetScopes(), ExpandScopes(s.Scopes...)...)
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)
	switch r := req.(type) {
	case *RefreshRequest:
		r.SetCurrentScopes(scopes)
	case *AuthRequest:
		r.Scopes = scopes
	case *ClientRequest:
		r.scopes = scopes
	case *op.DeviceAuthorizationState:
		r.Scopes = scopes
	}
	return
}

// createRefreshToken creates a refresh token and updates the grant.
func (r *Storage) createRefreshToken(ctx context.Context, req op.TokenRequest) (tokenId string, err error) {
	authReq, cast := req.(op.AuthRequest)
	if !cast {
		return
	}
	refreshToken := r.genId()
	refreshTokenHash := secret.Hash(refreshToken)
	grant := &Grant{}
	err = r.db.First(grant, "authId = ?", authReq.GetID()).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	grant.RefreshToken = refreshTokenHash
	err = r.db.Save(grant).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	authCode := r.authCodeById(authReq.GetID())
	if authCode != "" {
		// Auth code flow - delete auth request
		err = r.deleteAuthRequestByCode(ctx, authCode)
		if err != nil {
			return
		}
	} else {
		// Device flow - delete device authorization
		deviceCode := r.devAuthCodeBySubject(authReq.GetSubject())
		if deviceCode != "" {
			r.DeleteDevAuthByCode(deviceCode)
		}
	}
	tokenId = refreshToken
	return
}

// refreshIdentity refreshes the Idp identity.
func (r *Storage) refreshIdentity(req op.RefreshTokenRequest, grant *Grant) (err error) {
	subject := req.GetSubject()
	s, err := r.findSubject(subject)
	if err != nil {
		if errors.Is(err, &NotFound{}) {
			err = nil
		}
		return
	}
	if !s.IsIdentity() {
		// Local user - nothing to refresh
		return
	}
	// Refresh based on identity kind
	switch s.Identity.Kind {
	case IdentityKindLDAP:
		err = r.refreshLdapIdentity(s.Identity, grant)
		if err != nil {
			return
		}
		_, err = secret.Encode(grant)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = r.db.Save(grant).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case IdentityKindOpenid:
		login := &FedIdpLogin{
			handler:  r.idpHandler,
			identity: s.Identity,
		}
		err = login.refreshIdentity(grant)
		if err != nil {
			return
		}
		_, err = secret.Encode(grant)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = r.db.Save(grant).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// refreshLdapIdentity re-authenticates with LDAP and updates the identity.
func (r *Storage) refreshLdapIdentity(identity *Identity, grant *Grant) (err error) {
	if !r.dsHandler.enabled {
		return
	}
	password := grant.IdpRefreshToken
	subject, err := r.dsHandler.Authenticate(identity.Login, password)
	if err != nil {
		return
	}
	if subject.Identity != nil {
		*identity = *subject.Identity
	}
	return
}

// token returns a token by id.
func (r *Storage) token(_ context.Context, id string) (m *Token, err error) {
	m = &Token{}
	err = r.db.First(m, "authId", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = oidc.ErrInvalidGrant().WithDescription("token not-found.")
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	return
}

// deleteToken deletes a token by id.
func (r *Storage) deleteToken(_ context.Context, id string) (err error) {
	var tokens []*Token
	db := r.db.Where("authId", id)
	db = db.Where("kind", KindAccessToken)
	err = db.Find(&tokens).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range tokens {
		err = r.db.Delete(m).Error
		if err == nil {
			r.cache.TokenDeleted(m.ID)
		} else {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// deleteTokensBySubject deletes all tokens for a subject.
func (r *Storage) deleteTokensBySubject(_ context.Context, subject string) (err error) {
	var tokens []*Token
	db := r.db.Where("subject", subject)
	db = db.Where("kind", KindAccessToken)
	err = db.Find(&tokens).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range tokens {
		err = r.db.Delete(m).Error
		if err == nil {
			r.cache.TokenDeleted(m.ID)
		} else {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// grantByRefreshToken returns a grant by refresh token.
func (r *Storage) grantByRefreshToken(_ context.Context, token string) (m *Grant, err error) {
	m = &Grant{}
	db := r.db.Where("expiration > ?", time.Now())
	db = db.Where("refreshToken", secret.Hash(token))
	err = db.First(m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = oidc.ErrInvalidGrant().WithDescription("grant not-found.")
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	err = secret.Decode(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.orphaned(m)
	if err != nil {
		return
	}
	return
}

// deleteGrant deletes a grant by id.
func (r *Storage) deleteGrant(_ context.Context, id string) (err error) {
	m := &Grant{}
	err = r.db.Delete(m, "authId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// orphaned imposes grant expiration when the user cannot be found.
func (r *Storage) orphaned(grant *Grant) (err error) {
	if grant.Kind != KindAuthCode {
		return
	}
	_, err = r.cache.FindSubject(grant.Subject)
	if err != nil {
		if errors.Is(err, &NotFound{}) {
			grant.Expiration = time.Now()
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	return
}

// createGrant creates a grant for an authorization request.
// The authId parameter is used as the grant ID and for linking tokens.
// The authCode parameter is optional (empty string for device flow).
// A temporary refresh token is generated to avoid UNIQUE constraint race conditions;
// it will be replaced by createRefreshToken() for auth code flows.
func (r *Storage) createGrant(
	_ context.Context,
	req GrantRequest,
	authId string,
	authCode string) (grantId string, err error) {
	//
	grantId = authId
	expiration := time.Now().Add(Settings.Token.RefreshLifespan)

	m := &Grant{}
	m.Kind = KindAuthCode
	m.ClientId = req.GetClientID()
	m.AuthId = grantId
	m.Subject = req.GetSubject()
	m.AuthCode = authCode
	m.RefreshToken = r.genId()
	m.Scopes = req.GetScopes()
	m.Issued = req.GetAuthTime()
	m.Expiration = expiration
	subject, err := r.findSubject(req.GetSubject())
	if err != nil {
		return
	}
	m.UserID = subject.UserId
	m.IdpIdentityID = subject.IdentityId
	m.IdpClientID = subject.ClientId
	if authReq, cast := req.(*AuthRequest); cast {
		if authReq.idpRefreshToken != "" {
			m.IdpRefreshToken = authReq.idpRefreshToken
		}
	}
	_, err = secret.Encode(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// authCodeById returns the auth code for an auth request.
func (r *Storage) authCodeById(id string) (code string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	authReq, found := r.authReqById[id]
	if found {
		code = authReq.authCode
	}
	return
}

// deleteAuthRequestByCode deletes auth request by code.
func (r *Storage) deleteAuthRequestByCode(_ context.Context, code string) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	requestId, found := r.authReqByCode[code]
	if !found {
		return
	}
	delete(r.authReqByCode, code)
	delete(r.authReqById, requestId)
	return
}

// findSubject finds and resolves a subject to User or IdpIdentity.
func (r *Storage) findSubject(subject string) (s *Subject, err error) {
	s, err = r.cache.FindSubject(subject)
	return
}

// grantId returns the grant ID by authId.
func (r *Storage) grantId(authId string) (id *uint) {
	m := &Grant{}
	err := r.db.First(m, "authId", authId).Error
	if err != nil {
		return
	}
	id = &m.ID
	return
}

// grantId returns the grant authID by id.
func (r *Storage) grantAuthId(id uint) (authId string) {
	m := &Grant{}
	err := r.db.First(m, id).Error
	if err != nil {
		return
	}
	authId = m.AuthId
	return
}

// genId returns a new generated ID.
func (r *Storage) genId() (s string) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	s = hex.EncodeToString(b)
	return
}

// StoreDeviceAuthorization stores a new device authorization request.
func (r *Storage) StoreDeviceAuthorization(
	_ context.Context,
	clientId, deviceCode, userCode string,
	expires time.Time,
	scopes []string) (err error) {
	//
	now := time.Now()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for devCode, req := range r.devAuthReqByDevCode {
		if now.After(req.expiration) {
			delete(r.devAuthReqByDevCode, devCode)
			delete(r.devAuthByUserCode, req.userCode)
		}
	}
	devAuth := &DeviceAuthRequest{
		deviceCode: deviceCode,
		userCode:   userCode,
		clientId:   clientId,
		scopes:     scopes,
		issued:     time.Now(),
		expiration: expires,
	}
	r.devAuthReqByDevCode[deviceCode] = devAuth
	r.devAuthByUserCode[userCode] = deviceCode
	return
}

// GetDeviceAuthorizatonState returns device authorization state.
// Note: Method name is intentionally misspelled per zitadel API.
func (r *Storage) GetDeviceAuthorizatonState(
	_ context.Context,
	clientId, deviceCode string) (state *op.DeviceAuthorizationState, err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	r.mutex.Lock()
	devAuth, found := r.devAuthReqByDevCode[deviceCode]
	r.mutex.Unlock()

	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("deviceCode not-found.")
		return
	}

	state = &op.DeviceAuthorizationState{
		ClientID: clientId,
		Scopes:   devAuth.scopes,
		Expires:  devAuth.expiration,
		Done:     devAuth.done,
		Denied:   devAuth.denied,
		Subject:  devAuth.subject,
		AuthTime: devAuth.authTime,
	}
	return
}

// DeleteDevAuthByCode deletes device authorization by device code.
func (r *Storage) DeleteDevAuthByCode(deviceCode string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	devAuth, found := r.devAuthReqByDevCode[deviceCode]
	if found {
		delete(r.devAuthByUserCode, devAuth.userCode)
		delete(r.devAuthReqByDevCode, deviceCode)
	}
}

// devAuthCodeBySubject returns device code for completed device authorization by subject.
func (r *Storage) devAuthCodeBySubject(subject string) (deviceCode string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for code, req := range r.devAuthReqByDevCode {
		if req.subject == subject && req.done {
			deviceCode = code
			return
		}
	}
	return
}

// GetDevAuthByUserCode returns device authorization by user code.
// Returns the device authorization request and true if found, otherwise nil and false.
func (r *Storage) GetDevAuthByUserCode(userCode string) (devAuth *DeviceAuthRequest, found bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	deviceCode, found := r.devAuthByUserCode[userCode]
	if !found {
		return
	}
	devAuth, found = r.devAuthReqByDevCode[deviceCode]
	return
}

// UpdateDevAuth updates device authorization state.
// Sets the subject, completion status, denial status, and authorization time
// for the device authorization identified by the user code.
func (r *Storage) UpdateDevAuth(userCode string, subject string, done bool, denied bool, authTime time.Time) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	deviceCode, found := r.devAuthByUserCode[userCode]
	if !found {
		err = fmt.Errorf("device authorization not found")
		return
	}
	devAuth, found := r.devAuthReqByDevCode[deviceCode]
	if !found {
		err = fmt.Errorf("device authorization not found")
		return
	}
	devAuth.subject = subject
	devAuth.done = done
	devAuth.denied = denied
	devAuth.authTime = authTime
	return
}

// Client implements op.Client.
type Client struct {
	id              string
	subject         string
	secret          string
	redirectURIs    []string
	grantTypes      []string
	applicationType op.ApplicationType
	scopes          []string
	request         *http.Request
}

// GetID returns the client ID.
func (c *Client) GetID() (s string) {
	s = c.id
	return
}

// RedirectURIs returns redirect URIs.
func (c *Client) RedirectURIs() (uris []string) {
	uris = c.redirectURIs
	return
}

// PostLogoutRedirectURIs returns post-logout redirect URIs.
func (c *Client) PostLogoutRedirectURIs() (uris []string) {
	uris = c.redirectURIs
	return
}

// ApplicationType returns the application type.
func (c *Client) ApplicationType() (t op.ApplicationType) {
	t = c.applicationType
	return
}

// AuthMethod returns the authentication method.
func (c *Client) AuthMethod() (m oidc.AuthMethod) {
	if c.secret != "" {
		m = oidc.AuthMethodPost
	} else {
		m = oidc.AuthMethodNone
	}
	return
}

// ResponseTypes returns response types.
func (c *Client) ResponseTypes() (types []oidc.ResponseType) {
	types = []oidc.ResponseType{oidc.ResponseTypeCode}
	return
}

// GrantTypes returns grant types.
func (c *Client) GrantTypes() (types []oidc.GrantType) {
	for _, g := range c.grantTypes {
		types = append(types, oidc.GrantType(g))
	}
	return
}

// LoginURL returns the login URL.
func (c *Client) LoginURL(id string) (s string) {
	s = fmt.Sprintf(
		"%s?%s=%s",
		AppendIssuer(c.request, api.LoginRoute),
		AuthRequestId,
		id)
	return
}

// AccessTokenType returns the access token type.
func (c *Client) AccessTokenType() (t op.AccessTokenType) {
	t = op.AccessTokenTypeJWT
	return
}

// IDTokenLifetime returns the ID token lifetime.
func (c *Client) IDTokenLifetime() (d time.Duration) {
	d = Settings.Token.Lifespan
	return
}

// DevMode returns whether dev mode is enabled.
func (c *Client) DevMode() (b bool) {
	return
}

// RestrictAdditionalIdTokenScopes returns a scope restriction function.
func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

// RestrictAdditionalAccessTokenScopes returns a scope restriction function.
func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

// IsScopeAllowed checks if a scope is allowed.
func (c *Client) IsScopeAllowed(scope string) (b bool) {
	for i := range c.scopes {
		if c.scopes[i] == scope {
			b = true
			break
		}
	}
	return
}

// IDTokenUserinfoClaimsAssertion returns userinfo claims assertion setting.
func (c *Client) IDTokenUserinfoClaimsAssertion() (b bool) {
	return
}

// ClockSkew returns the clock skew.
func (c *Client) ClockSkew() (d time.Duration) {
	d = time.Minute
	return
}

// With populates op.Client using the model.
func (c *Client) With(m *IdpClient, req *http.Request) {
	c.id = m.ClientId
	c.subject = m.Subject
	c.secret = m.Secret
	c.grantTypes = m.Grants
	c.redirectURIs = m.RedirectURIs
	c.scopes = m.Scopes
	c.request = req
	switch m.ApplicationType {
	case "web":
		c.applicationType = op.ApplicationTypeWeb
	case "native":
		c.applicationType = op.ApplicationTypeNative
	}

}

// Inject template values:
// - * wildcard matched against the requested URI.
// - ${issuer}
// - ${issuer.proto}
// - ${issuer.host}
// - ${issuer.port}
// - ${issuer.path}
func (c *Client) Inject() {
	req := c.request
	issuer := Issuer(req)
	requested := c.requestedURI()
	issuerURL, _ := url.Parse(issuer)
	for i, u := range c.redirectURIs {
		u = strings.Replace(u, "${issuer}", issuer, -1)
		u = strings.Replace(u, "${issuer.proto}", issuerURL.Scheme, -1)
		u = strings.Replace(u, "${issuer.host}", issuerURL.Hostname(), -1)
		u = strings.Replace(u, "${issuer.port}", issuerURL.Port(), -1)
		u = strings.Replace(u, "${issuer.path}", issuerURL.Path, -1)
		if strings.ContainsAny(u, "*${}") {
			matched, _ := doublestar.Match(u, requested)
			if matched {
				u = requested
			}
		}
		c.redirectURIs[i] = u
	}
}

// requestedURI returns the redirect URI from the request query parameters.
// Returns redirect_uri (RFC 6749 §3.1.2) for authorization flows or
// post_logout_redirect_uri for logout flows (OpenID Connect RP-Initiated Logout 1.0 §2).
func (c *Client) requestedURI() (u string) {
	q := c.request.URL.Query()
	p := "redirect_uri"
	if strings.Contains(c.request.URL.Path, "/logout") {
		p = "post_logout_" + p
	}
	u = q.Get(p)
	return
}

// GrantRequest defines the interface needed for creating grants.
type GrantRequest interface {
	GetClientID() string
	GetSubject() string
	GetScopes() []string
	GetAuthTime() time.Time
}

// AuthRequest implements op.AuthRequest.
type AuthRequest struct {
	*oidc.AuthRequest
	requestId       string
	subject         string
	authCode        string
	issued          time.Time
	expiration      time.Time
	done            bool
	idpRefreshToken string
	idpIdentityId   uint
}

// GetID returns the request ID.
func (a *AuthRequest) GetID() (s string) {
	s = a.requestId
	return
}

// GetACR returns the ACR.
func (a *AuthRequest) GetACR() (s string) {
	s = ""
	return
}

// GetAMR returns the AMR.
func (a *AuthRequest) GetAMR() (amr []string) {
	amr = []string{"pwd"}
	return
}

// GetAudience returns the audience.
func (a *AuthRequest) GetAudience() (aud []string) {
	aud = []string{a.ClientID}
	return
}

// GetAuthTime returns the authentication time.
func (a *AuthRequest) GetAuthTime() (t time.Time) {
	t = a.issued
	return
}

// GetClientID returns the client ID.
func (a *AuthRequest) GetClientID() (s string) {
	s = a.ClientID
	return
}

// GetCodeChallenge returns the code challenge.
func (a *AuthRequest) GetCodeChallenge() (challenge *oidc.CodeChallenge) {
	if a.CodeChallenge != "" {
		challenge = &oidc.CodeChallenge{
			Challenge: a.CodeChallenge,
			Method:    a.CodeChallengeMethod,
		}
	}
	return
}

// GetNonce returns the nonce.
func (a *AuthRequest) GetNonce() (s string) {
	s = a.Nonce
	return
}

// GetRedirectURI returns the redirect URI.
func (a *AuthRequest) GetRedirectURI() (s string) {
	s = a.RedirectURI
	return
}

// GetResponseType returns the response type.
func (a *AuthRequest) GetResponseType() (t oidc.ResponseType) {
	if a.ResponseType != "" {
		t = a.ResponseType
	} else {
		t = oidc.ResponseTypeCode
	}
	return
}

// GetResponseMode returns the response mode.
func (a *AuthRequest) GetResponseMode() (m oidc.ResponseMode) {
	m = oidc.ResponseModeQuery
	return
}

// GetScopes returns the scopes.
func (a *AuthRequest) GetScopes() (scopes []string) {
	if len(a.Scopes) > 0 {
		scopes = a.Scopes
	} else {
		scopes = []string{"openid"}
	}
	return
}

// GetState returns the state.
func (a *AuthRequest) GetState() (s string) {
	s = a.State
	return
}

// GetSubject returns the subject.
func (a *AuthRequest) GetSubject() (s string) {
	s = a.subject
	return
}

// Done returns whether the request is done.
func (a *AuthRequest) Done() (b bool) {
	b = a.done
	return
}

// RefreshRequest implements op.RefreshTokenRequest.
type RefreshRequest struct {
	grantId  string
	clientId string
	subject  string
	scopes   []string
	issued   time.Time
}

// GetAMR returns the AMR.
func (r *RefreshRequest) GetAMR() (amr []string) {
	amr = []string{"pwd"}
	return
}

// GetAudience returns the audience.
func (r *RefreshRequest) GetAudience() (aud []string) {
	aud = []string{r.clientId}
	return
}

// GetAuthTime returns the authentication time.
func (r *RefreshRequest) GetAuthTime() (t time.Time) {
	t = r.issued
	return
}

// GetClientID returns the client ID.
func (r *RefreshRequest) GetClientID() (s string) {
	s = r.clientId
	return
}

// GetScopes returns the scopes.
func (r *RefreshRequest) GetScopes() (scopes []string) {
	scopes = r.scopes
	return
}

// GetSubject returns the subject.
func (r *RefreshRequest) GetSubject() (s string) {
	s = r.subject
	return
}

// SetCurrentScopes sets the current scopes.
func (r *RefreshRequest) SetCurrentScopes(scopes []string) {
	r.scopes = scopes
	return
}

// ClientRequest implements op.TokenRequest for client credentials flow.
type ClientRequest struct {
	authId   string
	clientId string
	subject  string
	scopes   []string
	issued   time.Time
}

// GetAMR returns the AMR.
func (r *ClientRequest) GetAMR() (amr []string) {
	amr = []string{"client_credentials"}
	return
}

// GetAudience returns the audience.
func (r *ClientRequest) GetAudience() (aud []string) {
	aud = []string{r.clientId}
	return
}

// GetAuthTime returns the authentication time.
func (r *ClientRequest) GetAuthTime() (t time.Time) {
	t = r.issued
	return
}

// GetClientID returns the client ID.
func (r *ClientRequest) GetClientID() (s string) {
	s = r.clientId
	return
}

// GetScopes returns the scopes.
func (r *ClientRequest) GetScopes() (scopes []string) {
	scopes = r.scopes
	return
}

// GetSubject returns the subject.
func (r *ClientRequest) GetSubject() (s string) {
	s = r.subject
	return
}

// DeviceAuthRequest holds device authorization state.
type DeviceAuthRequest struct {
	deviceCode string
	userCode   string
	clientId   string
	subject    string
	scopes     []string
	issued     time.Time
	expiration time.Time
	done       bool
	denied     bool
	authTime   time.Time
}

// Done returns done status.
func (r *DeviceAuthRequest) Done() (done bool) {
	done = r.done
	return
}

// Denied returns denied status.
func (r *DeviceAuthRequest) Denied() (denied bool) {
	denied = r.denied
	return
}

// Expiration returns expiration time.
func (r *DeviceAuthRequest) Expiration() (exp time.Time) {
	exp = r.expiration
	return
}

// Key implements op.Key.
type Key struct {
	jwk JWK
}

// Algorithm returns the signature algorithm.
func (k *Key) Algorithm() (s jose.SignatureAlgorithm) {
	s = jose.SignatureAlgorithm(k.jwk.Algorithm)
	return
}

// Use returns the key use.
func (k *Key) Use() (s string) {
	s = k.jwk.Use
	return
}

// Key returns the public key for verification.
func (k *Key) Key() (key any) {
	if rsaKey, cast := k.jwk.PrivateKey.(*rsa.PrivateKey); cast {
		key = &rsaKey.PublicKey
	} else {
		key = k.jwk.PrivateKey
	}
	return
}

// ID returns the key ID.
func (k *Key) ID() (s string) {
	s = k.jwk.KeyID
	return
}
