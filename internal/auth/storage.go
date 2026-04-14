package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TokenTypeRefresh  = "refresh_token"
	TokenTypeAuthCode = "authorization_code"
)

// Storage implements op.Storage for zitadel/oidc.
type Storage struct {
	keySet     KeySet
	db         *gorm.DB
	authReqs   map[string]*AuthRequest
	authByCode map[string]string
	mu         sync.RWMutex
}

// GetClientByClientID retrieves a client by ID.
func (r *Storage) GetClientByClientID(_ context.Context, clientId string) (client op.Client, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "client lookup failed", "clientID", clientId)
		}
	}()
	c := &Client{
		id:           Settings.Auth.Client.ID,
		secret:       Settings.Auth.Client.Secret,
		redirectURIs: r.redirectURIs(),
	}
	if clientId != c.id {
		err = oidc.ErrInvalidClient().WithDescription("client not found")
		return
	}
	client = c
	return
}

// AuthorizeClientIDSecret validates client credentials.
func (r *Storage) AuthorizeClientIDSecret(
	ctx context.Context,
	clientID, clientSecret string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "client auth failed")
		}
	}()
	client, err := r.GetClientByClientID(ctx, clientID)
	if err != nil {
		return
	}
	c := client.(*Client)
	if c.secret != "" && c.secret != clientSecret {
		err = oidc.ErrInvalidClient().WithDescription("invalid client secret")
	}
	return
}

// ClientCredentials validates client credentials for client credentials flow.
func (r *Storage) ClientCredentials(
	ctx context.Context,
	clientID, clientSecret string) (client op.Client, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "client credentials validation failed")
		}
	}()
	client, err = r.GetClientByClientID(ctx, clientID)
	if err != nil {
		return
	}
	c := client.(*Client)
	if c.secret != "" && c.secret != clientSecret {
		err = oidc.ErrInvalidClient().WithDescription("invalid client secret")
		client = nil
		return
	}
	return
}

// ClientCredentialsTokenRequest creates a token request for client credentials.
func (r *Storage) ClientCredentialsTokenRequest(
	ctx context.Context,
	clientID string,
	scopes []string) (req op.TokenRequest, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "client credentials token request failed")
		}
	}()
	req = &TokenRequest{
		grantID:  r.genId(),
		clientID: clientID,
		subject:  clientID,
		scopes:   scopes,
	}
	return
}

// CreateAuthRequest initiates an authorization request.
func (r *Storage) CreateAuthRequest(
	ctx context.Context,
	authReq *oidc.AuthRequest,
	userID string) (req op.AuthRequest, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "create auth request failed")
		}
	}()
	requestID := r.genId()
	req = &AuthRequest{
		AuthRequest: *authReq,
		RequestID:   requestID,
		Subject:     userID,
		AuthTime:    time.Now(),
		Expiration:  time.Now().Add(10 * time.Minute),
	}
	r.mu.Lock()
	r.authReqs[requestID] = req.(*AuthRequest)
	r.mu.Unlock()
	return
}

// AuthRequestByID retrieves an auth request by ID.
func (r *Storage) AuthRequestByID(
	ctx context.Context,
	id string) (req op.AuthRequest, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "auth request lookup failed", "id", id)
		}
	}()
	r.mu.RLock()
	req, found := r.authReqs[id]
	r.mu.RUnlock()
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	return
}

// AuthRequestByCode retrieves auth request by authorization code.
func (r *Storage) AuthRequestByCode(
	ctx context.Context,
	code string) (req op.AuthRequest, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "auth request by code failed")
		}
	}()
	r.mu.RLock()
	requestID, found := r.authByCode[code]
	r.mu.RUnlock()
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth code not found")
		return
	}
	r.mu.RLock()
	req, found = r.authReqs[requestID]
	r.mu.RUnlock()
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	return
}

// SaveAuthCode stores the authorization code.
func (r *Storage) SaveAuthCode(ctx context.Context, id, code string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "save auth code failed")
		}
	}()
	r.mu.Lock()
	defer r.mu.Unlock()
	authReq, found := r.authReqs[id]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	authReq.AuthCode = code
	r.authByCode[code] = id
	return
}

// DeleteAuthRequest deletes an auth request.
func (r *Storage) DeleteAuthRequest(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "delete auth request failed")
		}
	}()
	r.mu.Lock()
	defer r.mu.Unlock()
	authReq, found := r.authReqs[id]
	if !found {
		return
	}
	if authReq.AuthCode != "" {
		delete(r.authByCode, authReq.AuthCode)
	}
	delete(r.authReqs, id)
	return
}

// CreateAccessToken creates an access token.
func (r *Storage) CreateAccessToken(
	ctx context.Context,
	req op.TokenRequest) (tokenID string, expiration time.Time, err error) {
	expiration = time.Now().Add(time.Duration(Settings.Token.Lifespan) * time.Second)
	tokenID = r.genId()
	return
}

// CreateAccessAndRefreshTokens creates both access and refresh tokens.
func (r *Storage) CreateAccessAndRefreshTokens(
	ctx context.Context,
	req op.TokenRequest,
	currentRefreshToken string) (
	accessTokenID, newRefreshToken string,
	expiration time.Time,
	err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "create tokens failed")
		}
	}()
	accessTokenID, expiration, err = r.CreateAccessToken(ctx, req)
	if err != nil {
		return
	}
	newRefreshToken, err = r.createRefreshToken(ctx, req)
	return
}

// TokenRequestByRefreshToken retrieves token request by refresh token.
func (r *Storage) TokenRequestByRefreshToken(
	ctx context.Context,
	refreshToken string) (req op.RefreshTokenRequest, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "token request by refresh failed")
		}
	}()
	grant, err := r.grantByRefreshToken(ctx, refreshToken)
	if err != nil {
		return
	}
	req = &TokenRequest{
		grantID:  grant.GrantId,
		clientID: grant.ClientId,
		subject:  grant.Subject,
		scopes:   strings.Fields(grant.Scopes),
	}
	return
}

// TerminateSession terminates a user session.
func (r *Storage) TerminateSession(
	ctx context.Context,
	userID, clientID string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "terminate session failed")
		}
	}()
	var grants []model.Grant
	err = r.db.Find(&grants, "subject", userID).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range grants {
		grant := &grants[i]
		err = r.deleteGrant(ctx, grant.GrantId)
		if err != nil {
			return
		}
		err = r.deleteTokensByGrantID(ctx, grant.GrantId)
		if err != nil {
			return
		}
	}
	return
}

// RevokeToken revokes a token.
func (r *Storage) RevokeToken(
	ctx context.Context,
	tokenOrTokenID, userID, clientID string) *oidc.Error {
	err := r.deleteToken(ctx, tokenOrTokenID)
	if err != nil {
		Log.Error(err, "revoke token failed")
		return oidc.ErrServerError()
	}
	return nil
}

// SigningKey returns the current signing key.
func (r *Storage) SigningKey(ctx context.Context) (op.SigningKey, error) {
	return r.keySet.SigningKey(), nil
}

// SignatureAlgorithms returns supported signature algorithms.
func (r *Storage) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{jose.RS256}, nil
}

// KeySet returns the JWKS.
func (r *Storage) KeySet(ctx context.Context) (keys []op.Key, err error) {
	keys = make([]op.Key, 0)
	for _, jwk := range r.keySet.Keys {
		keys = append(keys, &Key{jwk: jwk})
	}
	return
}

// GetKeyByIDAndClientID retrieves a key.
func (r *Storage) GetKeyByIDAndClientID(
	ctx context.Context,
	keyID, clientID string) (key *jose.JSONWebKey, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "key lookup failed")
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
func (r *Storage) ValidateJWTProfileScopes(
	ctx context.Context,
	userID string,
	scopes []string) ([]string, error) {
	return scopes, nil
}

// Health checks storage health.
func (r *Storage) Health(ctx context.Context) (err error) {
	err = r.db.Exec("SELECT 1").Error
	return
}

// GetPrivateClaimsFromScopes returns private claims based on scopes.
func (r *Storage) GetPrivateClaimsFromScopes(
	ctx context.Context,
	userID, clientID string,
	scopes []string) (claims map[string]any, err error) {
	//
	claims = make(map[string]any)
	if userID == "" {
		return
	}
	user := &model.User{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("Roles.Permissions")
	err = db.First(user, "subject", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	roleNames := make([]string, 0, len(user.Roles))
	permissions := make([]string, 0)
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission.Scope)
		}
	}
	if len(roleNames) > 0 {
		claims["roles"] = roleNames
	}
	if len(permissions) > 0 {
		claims["permissions"] = permissions
	}
	return
}

// SetUserinfoFromScopes is deprecated but required by the interface.
func (r *Storage) SetUserinfoFromScopes(
	ctx context.Context,
	userinfo *oidc.UserInfo,
	userID, clientID string,
	scopes []string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "set userinfo from scopes failed")
		}
	}()
	user := &model.User{}
	err = r.db.First(user, "subject", userID).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	userinfo.Subject = user.Subject
	userinfo.PreferredUsername = user.Userid
	if user.Email != "" {
		userinfo.Email = user.Email
		userinfo.EmailVerified = oidc.Bool(true)
	}
	return
}

// SetUserinfoFromToken sets userinfo claims.
func (r *Storage) SetUserinfoFromToken(
	ctx context.Context,
	userinfo *oidc.UserInfo,
	tokenID, subject, origin string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "set userinfo failed")
		}
	}()
	user := &model.User{}
	err = r.db.First(user, "subject", subject).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	userinfo.Subject = user.Subject
	userinfo.PreferredUsername = user.Userid
	if user.Email != "" {
		userinfo.Email = user.Email
		userinfo.EmailVerified = oidc.Bool(true)
	}
	return
}

// SetIntrospectionFromToken sets introspection response.
func (r *Storage) SetIntrospectionFromToken(
	ctx context.Context,
	introspection *oidc.IntrospectionResponse,
	tokenID, subject, clientID string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "set introspection failed")
		}
	}()
	token, err := r.token(ctx, tokenID)
	if err != nil {
		return
	}
	expiration := int(token.Expiration.Unix())
	introspection.Active = expiration > int(time.Now().Unix())
	introspection.Scope = strings.Fields(token.Scopes)
	introspection.ClientID = token.ClientId
	introspection.Subject = token.Subject
	introspection.Expiration = oidc.FromTime(token.Expiration)
	return
}

// GetRefreshTokenInfo retrieves refresh token info.
func (r *Storage) GetRefreshTokenInfo(
	ctx context.Context,
	clientID, token string) (userID, tokenID string, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "get refresh token info failed")
		}
	}()
	grant, err := r.grantByRefreshToken(ctx, token)
	if err != nil {
		return
	}
	userID = grant.Subject
	tokenID = grant.RefreshToken
	return
}

// Login handles the authentication flow.
func (r *Storage) Login(
	writer http.ResponseWriter,
	request *http.Request,
	authReqID string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "login failed")
		}
	}()
	err = request.ParseForm()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	userid := request.PostFormValue("userid")
	password := request.PostFormValue("password")
	if userid == "" || password == "" {
		err = r.renderPage(writer, request, authReqID)
		return
	}
	user := &model.User{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("Roles.Permissions")
	err = db.First(user, "userid", userid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.renderPage(writer, request, authReqID)
			err = nil
		}
		return
	}
	if !secret.MatchPassword(password, user.Password) {
		err = r.renderPage(writer, request, authReqID)
		return
	}
	r.mu.Lock()
	authReq, found := r.authReqs[authReqID]
	r.mu.Unlock()
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	r.mu.Lock()
	authReq.Subject = user.Subject
	authReq.AuthTime = time.Now()
	authReq.IsDone = true
	r.mu.Unlock()
	issuer := r.issuer()
	callbackURL := fmt.Sprintf("%s/authorize/callback?id=%s", issuer, authReqID)
	http.Redirect(writer, request, callbackURL, http.StatusFound)
	return
}

// renderPage renders the login page.
func (r *Storage) renderPage(
	writer http.ResponseWriter,
	_ *http.Request,
	authReqID string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "render page failed")
		}
	}()
	issuer := r.issuer()
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Tackle Hub - Login</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 400px;
            margin: 100px auto;
            padding: 20px;
        }
        h1 { color: #333; }
        form {
            background: #f5f5f5;
            padding: 20px;
            border-radius: 5px;
        }
        input {
            width: 100%;
            padding: 8px;
            margin: 10px 0;
            box-sizing: border-box;
        }
        button {
            background: #007bff;
            color: white;
            padding: 10px 20px;
            border: none;
            cursor: pointer;
            width: 100%;
        }
        button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <h1>Tackle Login</h1>
    <form action="` + issuer + `/login?authRequestID=` + authReqID + `" method="post">
        <div>
            <label>Userid:</label>
            <input type="text" name="userid" required autofocus />
        </div>
        <div>
            <label>Password:</label>
            <input type="password" name="password" required />
        </div>
        <button type="submit">Login</button>
    </form>
</body>
</html>`
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = writer.Write([]byte(html))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// createRefreshToken creates a refresh token.
func (r *Storage) createRefreshToken(ctx context.Context, req op.TokenRequest) (tokenID string, err error) {
	var authCode string
	authReq, cast := req.(op.AuthRequest)
	if !cast {
		return
	}
	tokenID = r.genId()
	expiration := time.Now().Add(
		time.Duration(Settings.Token.RefreshLifespan) * time.Second)
	userID, _ := r.userID(req.GetSubject())
	refreshToken := r.genId()
	digest := secret.Hash(refreshToken)
	grantID, err := r.createGrant(ctx, authReq, refreshToken, digest)
	if err != nil {
		return "", err
	}
	authCode = r.authCodeByID(authReq.GetID())
	m := &model.Token{
		TokenId:    tokenID,
		GrantId:    grantID,
		ClientId:   "",
		Subject:    req.GetSubject(),
		Type:       TokenTypeRefresh,
		Scopes:     strings.Join(req.GetScopes(), " "),
		Resources:  []string{},
		Issued:     time.Now(),
		Expiration: expiration,
		UserID:     userID,
	}
	err = r.db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return "", err
	}
	if authCode != "" {
		err = r.deleteAuthRequestByCode(ctx, authCode)
	}
	tokenID = refreshToken
	return
}

// token returns a token by id.
func (r *Storage) token(ctx context.Context, id string) (m *model.Token, err error) {
	m = &model.Token{}
	err = r.db.First(m, "tokenId", id).Error
	if err != nil {
		err = r.notFound(err)
		return
	}
	return
}

// deleteToken deletes a token by id.
func (r *Storage) deleteToken(ctx context.Context, id string) (err error) {
	m := &model.Token{}
	err = r.db.Delete(m, "tokenId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// deleteTokensByGrantID deletes tokens by grant id.
func (r *Storage) deleteTokensByGrantID(ctx context.Context, id string) (err error) {
	m := &model.Token{}
	err = r.db.Delete(m, "grantId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// grantByAuthCode returns a grant by auth code.
func (r *Storage) grantByAuthCode(
	ctx context.Context,
	code string) (m *model.Grant, err error) {
	m = &model.Grant{}
	err = r.db.First(m, "authCode", code).Error
	if err != nil {
		err = r.notFound(err)
		return
	}
	err = secret.Decrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// grantByRefreshToken returns a grant by refresh token.
func (r *Storage) grantByRefreshToken(
	ctx context.Context,
	token string) (m *model.Grant, err error) {
	digest := secret.Hash(token)
	m = &model.Grant{}
	err = r.db.First(m, "tokenDigest", digest).Error
	if err != nil {
		err = r.notFound(err)
		return
	}
	err = r.revoked(m)
	if err != nil {
		return
	}
	err = r.orphaned(m)
	if err != nil {
		return
	}
	err = secret.Decrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// deleteGrant deletes a grant by id.
func (r *Storage) deleteGrant(ctx context.Context, id string) (err error) {
	m := &model.Grant{}
	err = r.db.Delete(m, "grantId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// deleteGrantByAuthCode deletes a grant by auth code.
func (r *Storage) deleteGrantByAuthCode(ctx context.Context, code string) (err error) {
	m := &model.Grant{}
	err = r.db.Delete(m, "authCode", code).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// revoked enforces token revocation.
func (r *Storage) revoked(grant *model.Grant) (err error) {
	token, err := r.tokenByGrantID(grant.GrantId)
	if err != nil {
		return
	}
	if !token.Revoked.IsZero() {
		grant.Expiration = token.Revoked
		err = r.db.Save(grant).Error
	}
	return
}

// orphaned imposes grant expiration when the user cannot be found.
func (r *Storage) orphaned(grant *model.Grant) (err error) {
	if grant.Type != TokenTypeAuthCode {
		return
	}
	count := int64(0)
	user := &model.User{}
	db := r.db.Model(user)
	db = db.Where("subject", grant.Subject)
	err = db.Count(&count).Error
	if err != nil {
		err = r.notFound(err)
		return
	}
	if count == 0 {
		grant.Expiration = time.Now().UTC()
	}
	return
}

// tokenByGrantID returns a token by grant id.
func (r *Storage) tokenByGrantID(grantID string) (m *model.Token, err error) {
	m = &model.Token{}
	err = r.db.First(m, "grantId", grantID).Error
	if err != nil {
		err = r.notFound(err)
		return
	}
	return
}

// createGrant creates a grant from an auth request.
func (r *Storage) createGrant(
	ctx context.Context,
	authReq op.AuthRequest,
	refreshToken, digest string) (grantID string, err error) {
	grantID = r.genId()
	authCode := r.authCodeByID(authReq.GetID())
	err = r.createGrantDirect(
		ctx,
		grantID,
		authReq.GetClientID(),
		authReq.GetSubject(),
		authCode,
		authReq.GetScopes(),
		refreshToken,
		digest)
	return
}

// createGrantDirect creates a grant with explicit parameters.
func (r *Storage) createGrantDirect(
	ctx context.Context,
	grantID, clientID, subject, authCode string,
	scopes []string,
	refreshToken, digest string) (err error) {
	m := &model.Grant{
		GrantId:      grantID,
		ClientId:     clientID,
		Subject:      subject,
		TokenDigest:  digest,
		RefreshToken: refreshToken,
		AuthCode:     authCode,
		Type:         TokenTypeAuthCode,
		Scopes:       strings.Join(scopes, " "),
		Resources:    []string{},
		Expiration:   time.Now().Add(time.Duration(Settings.Token.RefreshLifespan) * time.Second),
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// authCodeByID returns the auth code for an auth request.
func (r *Storage) authCodeByID(id string) (code string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	authReq, found := r.authReqs[id]
	if found {
		code = authReq.AuthCode
	}
	return
}

// deleteAuthRequestByCode deletes auth request by code.
func (r *Storage) deleteAuthRequestByCode(ctx context.Context, code string) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	requestID, found := r.authByCode[code]
	if !found {
		return
	}
	delete(r.authByCode, code)
	delete(r.authReqs, requestID)
	return
}

// userID returns the user ID for the subject.
func (r *Storage) userID(subject string) (id *uint, err error) {
	user := &model.User{}
	err = r.db.First(user, "subject", subject).Error
	if err != nil {
		return
	}
	id = &user.ID
	return
}

// redirectURIs returns configured redirect URIs.
func (r *Storage) redirectURIs() (uris []string) {
	uris = Settings.Auth.Client.RedirectURIs
	if len(uris) == 0 {
		issuer := r.issuer()
		uris = []string{issuer + "/callback"}
	}
	return
}

// issuer returns the issuer URL.
func (r *Storage) issuer() (s string) {
	s = Settings.IssuerURL
	if s == "" {
		s = Settings.Addon.Hub.URL + api.OIDCRoutes
	}
	return
}

// genId returns a new generated ID.
func (r *Storage) genId() (s string) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	s = base64.RawURLEncoding.EncodeToString(b)
	return
}

// notFound returns op-specific not found error.
// notFound maps gorm not found errors to OIDC InvalidGrant errors.
func (r *Storage) notFound(err error) (e2 error) {
	if err == nil {
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		e2 = oidc.ErrInvalidGrant().WithDescription("resource not found")
	} else {
		e2 = liberr.Wrap(err)
	}
	return
}

// Client implements op.Client.
type Client struct {
	id           string
	secret       string
	redirectURIs []string
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
	t = op.ApplicationTypeWeb
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
	types = []oidc.GrantType{
		oidc.GrantTypeCode,
		oidc.GrantTypeRefreshToken,
		oidc.GrantTypeClientCredentials,
		oidc.GrantTypeBearer,
	}
	return
}

// LoginURL returns the login URL.
func (c *Client) LoginURL(id string) (s string) {
	issuer := Settings.IssuerURL
	if issuer == "" {
		issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
	}
	s = fmt.Sprintf("%s/login?authRequestID=%s", issuer, id)
	return
}

// AccessTokenType returns the access token type.
func (c *Client) AccessTokenType() (t op.AccessTokenType) {
	t = op.AccessTokenTypeJWT
	return
}

// IDTokenLifetime returns the ID token lifetime.
func (c *Client) IDTokenLifetime() (d time.Duration) {
	d = time.Duration(Settings.Token.Lifespan) * time.Second
	return
}

// DevMode returns whether dev mode is enabled.
func (c *Client) DevMode() (b bool) {
	b = false
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
	b = true
	return
}

// IDTokenUserinfoClaimsAssertion returns userinfo claims assertion setting.
func (c *Client) IDTokenUserinfoClaimsAssertion() (b bool) {
	b = false
	return
}

// ClockSkew returns the clock skew.
func (c *Client) ClockSkew() (d time.Duration) {
	d = time.Minute
	return
}

// AuthRequest implements op.AuthRequest.
type AuthRequest struct {
	oidc.AuthRequest
	RequestID  string
	Subject    string
	AuthCode   string
	AuthTime   time.Time
	Expiration time.Time
	IsDone     bool
}

// GetID returns the request ID.
func (a *AuthRequest) GetID() (s string) {
	s = a.RequestID
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
	t = a.AuthTime
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
	s = a.Subject
	return
}

// Done returns whether the request is done.
func (a *AuthRequest) Done() (b bool) {
	b = a.IsDone
	return
}

// TokenRequest implements op.RefreshTokenRequest.
type TokenRequest struct {
	grantID  string
	clientID string
	subject  string
	scopes   []string
}

// GetAMR returns the AMR.
func (r *TokenRequest) GetAMR() (amr []string) {
	amr = []string{"pwd"}
	return
}

// GetAudience returns the audience.
func (r *TokenRequest) GetAudience() (aud []string) {
	aud = []string{r.clientID}
	return
}

// GetAuthTime returns the authentication time.
func (r *TokenRequest) GetAuthTime() (t time.Time) {
	t = time.Now()
	return
}

// GetClientID returns the client ID.
func (r *TokenRequest) GetClientID() (s string) {
	s = r.clientID
	return
}

// GetScopes returns the scopes.
func (r *TokenRequest) GetScopes() (scopes []string) {
	scopes = r.scopes
	return
}

// GetSubject returns the subject.
func (r *TokenRequest) GetSubject() (s string) {
	s = r.subject
	return
}

// SetCurrentScopes sets the current scopes.
func (r *TokenRequest) SetCurrentScopes(scopes []string) {
	r.scopes = scopes
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
