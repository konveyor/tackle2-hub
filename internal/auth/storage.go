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

// Storage implements op.Storage.
type Storage struct {
	mutex      sync.RWMutex
	keySet     KeySet
	db         *gorm.DB
	authReqs   map[string]*AuthRequest
	authByCode map[string]string
}

// Subject represents a resolved subject (User or IdpIdentity).
type Subject struct {
	name          string
	email         string
	roles         []string
	scopes        []string
	userId        *uint
	idpIdentityId *uint
}

// GetClientByClientID retrieves a client by ID.
func (r *Storage) GetClientByClientID(_ context.Context, clientId string) (client op.Client, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	found := &Client{
		id:           Settings.Auth.Client.ID,
		secret:       Settings.Auth.Client.Secret,
		redirectURIs: r.redirectURIs(),
	}
	if clientId != found.id {
		err = oidc.ErrInvalidClient().WithDescription("client not found")
		return
	}
	client = found
	return
}

// AuthorizeClientIDSecret validates client credentials.
func (r *Storage) AuthorizeClientIDSecret(ctx context.Context, id, secret string) (err error) {
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
	if found.secret != secret {
		err = oidc.ErrInvalidClient().WithDescription("invalid client secret")
	}
	return
}

// ClientCredentials validates client credentials for client credentials flow.
func (r *Storage) ClientCredentials(ctx context.Context, id, secret string) (client op.Client, err error) {
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
	if found.secret != secret {
		err = oidc.ErrInvalidClient().WithDescription("invalid client secret")
	}
	return
}

// ClientCredentialsTokenRequest creates a token request for client credentials.
func (r *Storage) ClientCredentialsTokenRequest(
	_ context.Context,
	clientId string,
	scopes []string) (req op.TokenRequest, err error) {
	//
	req = &RefreshRequest{
		grantId:  r.genId(),
		clientId: clientId,
		subject:  clientId,
		scopes:   scopes,
	}
	return
}

// CreateAuthRequest initiates an authorization request.
func (r *Storage) CreateAuthRequest(
	_ context.Context,
	authReq *oidc.AuthRequest,
	userID string) (req op.AuthRequest, err error) {
	//
	r.mutex.Lock()
	defer r.mutex.Unlock()
	requestId := r.genId()
	req = &AuthRequest{
		AuthRequest: authReq,
		requestId:   requestId,
		subject:     userID,
		issued:      time.Now(),
		expiration:  time.Now().Add(10 * time.Minute),
	}
	r.authReqs[requestId] = req.(*AuthRequest)
	return
}

// AuthRequestByID retrieves an auth request by ID.
func (r *Storage) AuthRequestByID(_ context.Context, id string) (req op.AuthRequest, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	req, found := r.authReqs[id]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	return
}

// AuthRequestByCode retrieves auth request by authorization code.
func (r *Storage) AuthRequestByCode(_ context.Context, code string) (req op.AuthRequest, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	requestId, found := r.authByCode[code]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth code not found")
		return
	}
	req, found = r.authReqs[requestId]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
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
	authReq, found := r.authReqs[id]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	authReq.authCode = code
	r.authByCode[code] = id
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
	authReq, found := r.authReqs[id]
	if !found {
		return
	}
	delete(r.authByCode, authReq.authCode)
	delete(r.authReqs, id)
	return
}

// CreateAccessToken creates an access token.
func (r *Storage) CreateAccessToken(
	_ context.Context,
	req op.TokenRequest) (tokenId string, expiration time.Time, err error) {
	//
	err = r.injectScopes(req)
	if err != nil {
		return
	}
	tokenId = r.genId()
	subject := req.GetSubject()
	s, err := r.findSubject(subject)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
	}
	grantId := ""
	expiration = time.Now().Add(
		time.Duration(Settings.Token.Lifespan) * time.Second)
	switch r := req.(type) {
	case *RefreshRequest:
		grantId = r.grantId
	case *AuthRequest:
		//
	default:
		return
	}
	m := &model.Token{
		Kind:       KindAccessToken,
		AuthId:     tokenId,
		Subject:    subject,
		Scopes:     strings.Join(req.GetScopes(), " "),
		Issued:     time.Now(),
		Expiration: expiration,
		GrantID:    r.grantId(grantId),
	}
	if s != nil {
		m.UserID = s.userId
		m.IdpIdentityID = s.idpIdentityId
	}
	err = r.db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
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
			Log.Error(err, "")
		}
	}()
	grant, err := r.grantByRefreshToken(ctx, refreshToken)
	if err != nil {
		return
	}
	req = &RefreshRequest{
		grantId:  grant.AuthId,
		clientId: Settings.Auth.Client.ID,
		subject:  grant.Subject,
		scopes:   strings.Fields(grant.Scopes),
		issued:   grant.Issued,
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
	var grants []model.Grant
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
	grant := &model.Grant{}
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

	// Add roles
	if len(s.roles) > 0 {
		claims["roles"] = s.roles
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
	userinfo.PreferredUsername = s.name
	if s.email != "" {
		userinfo.Email = s.email
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
	userinfo.PreferredUsername = s.name
	if s.email != "" {
		userinfo.Email = s.email
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
	introspection.Scope = strings.Fields(token.Scopes)
	introspection.ClientID = Settings.Auth.Client.ID
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
	r.mutex.Lock()
	defer r.mutex.Unlock()
	defer func() {
		if err != nil {
			Log.Error(err, "")
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
		err = r.renderPage(writer, request, authReqId)
		return
	}
	user := &model.User{}
	err = r.db.First(user, "userid", userid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.renderPage(writer, request, authReqId)
			err = nil
		}
		return
	}
	if !secret.MatchPassword(password, user.Password) {
		err = r.renderPage(writer, request, authReqId)
		return
	}
	authReq, found := r.authReqs[authReqId]
	if !found {
		err = oidc.ErrInvalidGrant().WithDescription("auth request not found")
		return
	}
	authReq.subject = user.Subject
	authReq.issued = time.Now()
	authReq.done = true
	issuer := r.issuer()
	callbackURL := fmt.Sprintf("%s/authorize/callback?id=%s", issuer, authReqId)
	http.Redirect(writer, request, callbackURL, http.StatusFound)
	return
}

// renderPage renders the login page.
func (r *Storage) renderPage(writer http.ResponseWriter, _ *http.Request, authReqId string) (err error) {
	issuer := r.issuer()

	// Build external IdP button HTML if enabled
	idpButton := ""
	if Settings.Auth.Idp.Enabled {
		idpButton = `
        <div style="margin-top: 20px; text-align: center;">
            <div style="margin: 20px 0; color: #999;">- OR -</div>
            <a href="/idp/login?authRequestID=` + authReqId + `" style="
                display: block;
                background: #28a745;
                color: white;
                padding: 10px 20px;
                text-decoration: none;
                border-radius: 3px;
                text-align: center;
            ">Login with ` + Settings.Auth.Idp.Name + `</a>
        </div>`
	}

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
        a:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <h1>Tackle Login</h1>
    <form action="` + issuer + `/login?authRequestID=` + authReqId + `" method="post">
        <div>
            <label>Userid:</label>
            <input type="text" name="userid" required autofocus />
        </div>
        <div>
            <label>Password:</label>
            <input type="password" name="password" required />
        </div>
        <button type="submit">Login</button>
    </form>` + idpButton + `
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
	scopes := append(req.GetScopes(), s.scopes...)
	scopeMap := make(map[string]bool)
	uniqueScopes := make([]string, 0)
	for _, scope := range scopes {
		if !scopeMap[scope] {
			uniqueScopes = append(uniqueScopes, scope)
			scopeMap[scope] = true
		}
	}
	switch r := req.(type) {
	case *RefreshRequest:
		r.SetCurrentScopes(uniqueScopes)
	case *AuthRequest:
		r.Scopes = uniqueScopes
	default:
		//
	}
	return
}

// createRefreshToken creates a refresh token.
func (r *Storage) createRefreshToken(ctx context.Context, req op.TokenRequest) (tokenId string, err error) {
	authReq, cast := req.(op.AuthRequest)
	if !cast {
		return
	}
	refreshToken := r.genId()
	_, err = r.createGrant(ctx, authReq, refreshToken)
	if err != nil {
		return
	}
	authCode := r.authCodeById(authReq.GetID())
	if authCode != "" {
		err = r.deleteAuthRequestByCode(ctx, authCode)
		if err != nil {
			return
		}
	}
	tokenId = refreshToken
	return
}

// token returns a token by id.
func (r *Storage) token(_ context.Context, id string) (m *model.Token, err error) {
	m = &model.Token{}
	err = r.db.First(m, "authId", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = oidc.ErrInvalidGrant().WithDescription("token not found")
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	return
}

// deleteToken deletes a token by id.
func (r *Storage) deleteToken(_ context.Context, id string) (err error) {
	m := &model.Token{}
	err = r.db.Delete(m, "authId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// deleteTokensBySubject deletes all tokens for a subject.
func (r *Storage) deleteTokensBySubject(_ context.Context, subject string) (err error) {
	m := &model.Token{}
	err = r.db.Delete(m, "subject", subject).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// grantByRefreshToken returns a grant by refresh token.
func (r *Storage) grantByRefreshToken(_ context.Context, token string) (m *model.Grant, err error) {
	m = &model.Grant{}
	db := r.db.Where("expiration > ?", time.Now())
	db = db.Where("refreshToken", secret.Hash(token))
	err = db.First(m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = oidc.ErrInvalidGrant().WithDescription("grant not found")
		} else {
			err = liberr.Wrap(err)
		}
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
	m := &model.Grant{}
	err = r.db.Delete(m, "authId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// orphaned imposes grant expiration when the user cannot be found.
func (r *Storage) orphaned(grant *model.Grant) (err error) {
	if grant.Kind != KindAuthCode {
		return
	}
	count := int64(0)
	user := &model.User{}
	db := r.db.Model(user)
	db = db.Where("subject", grant.Subject)
	err = db.Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = oidc.ErrInvalidGrant().WithDescription("user not found")
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	if count == 0 {
		grant.Expiration = time.Now()
	}
	return
}

// createGrant creates a grant from an auth request.
func (r *Storage) createGrant(
	_ context.Context,
	authReq op.AuthRequest,
	refreshToken string) (grantId string, err error) {
	//
	grantId = r.genId()
	expiration := time.Now().
		Add(time.Duration(Settings.Token.RefreshLifespan) * time.Second)
	scopes := strings.Join(authReq.GetScopes(), " ")
	authCode := r.authCodeById(authReq.GetID())
	m := &model.Grant{
		Kind:         KindAuthCode,
		AuthId:       grantId,
		Subject:      authReq.GetSubject(),
		RefreshToken: secret.Hash(refreshToken),
		AuthCode:     authCode,
		Scopes:       scopes,
		Issued:       authReq.GetAuthTime(),
		Expiration:   expiration,
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
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	authReq, found := r.authReqs[id]
	if found {
		code = authReq.authCode
	}
	return
}

// deleteAuthRequestByCode deletes auth request by code.
func (r *Storage) deleteAuthRequestByCode(_ context.Context, code string) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	requestId, found := r.authByCode[code]
	if !found {
		return
	}
	delete(r.authByCode, code)
	delete(r.authReqs, requestId)
	return
}

// findSubject finds and resolves a subject to User or IdpIdentity.
func (r *Storage) findSubject(subject string) (s *Subject, err error) {
	user := &model.User{}
	db := r.db.Preload(clause.Associations).Preload("Roles.Permissions")
	err = db.First(user, "subject", subject).Error
	if err == nil {
		s = &Subject{
			name:   user.Userid,
			email:  user.Email,
			userId: &user.ID,
			roles:  make([]string, 0),
			scopes: make([]string, 0),
		}
		for _, role := range user.Roles {
			s.roles = append(s.roles, role.Name)
			// Extract permission scopes
			for _, permission := range role.Permissions {
				s.scopes = append(s.scopes, permission.Scope)
			}
		}
		return
	}
	idpIdentity := &model.IdpIdentity{}
	err = r.db.First(idpIdentity, "subject", subject).Error
	if err != nil {
		return
	}
	s = &Subject{
		name:          idpIdentity.Userid,
		email:         idpIdentity.Email,
		idpIdentityId: &idpIdentity.ID,
		roles:         make([]string, 0),
		scopes:        make([]string, 0),
	}
	if idpIdentity.Roles != "" {
		s.roles = strings.Fields(idpIdentity.Roles)
	}
	if idpIdentity.Scopes != "" {
		s.scopes = strings.Fields(idpIdentity.Scopes)
	}
	return
}

// grantId returns the grant ID by authId.
func (r *Storage) grantId(authId string) (id *uint) {
	m := &model.Grant{}
	err := r.db.First(m, "authId", authId).Error
	if err != nil {
		return
	}
	id = &m.ID
	return
}

// grantId returns the grant authID by id.
func (r *Storage) grantAuthId(id uint) (authId string) {
	m := &model.Grant{}
	err := r.db.First(m, id).Error
	if err != nil {
		return
	}
	authId = m.AuthId
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
	return
}

// ClockSkew returns the clock skew.
func (c *Client) ClockSkew() (d time.Duration) {
	d = time.Minute
	return
}

// AuthRequest implements op.AuthRequest.
type AuthRequest struct {
	*oidc.AuthRequest
	requestId  string
	subject    string
	authCode   string
	issued     time.Time
	expiration time.Time
	done       bool
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
