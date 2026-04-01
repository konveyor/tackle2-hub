package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"gorm.io/gorm"
)

// OIDC is the singleton auth provider.
var OIDC *BuiltinProvider

// KeySet alias.
type KeySet = goidc.JSONWebKeySet

// BuiltinProvider
type BuiltinProvider struct {
	openId *provider.Provider
	keySet KeySet
}

// Handler returns an http handler.
func (p *BuiltinProvider) Handler() http.Handler {
	return p.openId.Handler()
}

// Authenticate a web request.
func (p *BuiltinProvider) Authenticate(request *Request) (jwToken *jwt.Token, err error) {
	defer func() {
		if errors.Is(err, &NotValid{}) {
			Log.V(2).Info("[builtin] " + err.Error())
		}
	}()
	token, err := p.parseToken(request)
	if err != nil {
		return
	}
	jwToken, err = jwt.Parse(
		token,
		func(jwToken *jwt.Token) (key any, err error) {
			_, cast := jwToken.Method.(*jwt.SigningMethodRSA)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{Token: token})
				return
			}
			kid, found := jwToken.Header["kid"]
			if !found {
				err = liberr.Wrap(&NotAuthenticated{Token: token})
				return
			}
			key, err = p.keySet.Key(kid.(string))
			return
		})
	if err != nil {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	if !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	claims, cast := jwToken.Claims.(jwt.MapClaims)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Claims not specified.",
				Token:  token,
			})
		return
	}
	v, found := claims["sub"]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not specified.",
				Token:  token,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not string.",
				Token:  token,
			})
		return
	}
	v, found = claims["scope"]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not specified.",
				Token:  token,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not string.",
				Token:  token,
			})
		return
	}
	return
}

// Revoke an access token.
func (p *BuiltinProvider) Revoke(token *jwt.Token) (err error) {
	return
}

// Scopes returns a list of scopes.
func (p *BuiltinProvider) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	claims := jwToken.Claims.(jwt.MapClaims)
	for _, s := range strings.Fields(claims["scope"].(string)) {
		scope := &BaseScope{}
		scope.With(s)
		scopes = append(
			scopes,
			scope)
	}
	return
}

// parseToken returns the token
func (p *BuiltinProvider) parseToken(request *Request) (token string, err error) {
	splitToken := strings.Fields(request.Token)
	if len(splitToken) != 2 || strings.ToLower(splitToken[0]) != "bearer" {
		err = liberr.Wrap(&NotValid{Token: request.Token})
		return
	}
	token = splitToken[1]
	return
}

// New returns a configured provider.
func New(db *gorm.DB) (p *BuiltinProvider, err error) {
	p = &BuiltinProvider{}
	grantManager := NewGrantManager(db)
	keyManager := NewKeyManager(db)
	authManager := NewAuthManager(db)
	tokenManager := NewTokenManager(db)
	p.keySet, err = keyManager.KeySet()
	if err != nil {
		return
	}
	authPolicy := goidc.NewPolicy(
		"main",
		func(*http.Request, *goidc.Client, *goidc.AuthnSession) bool {
			return true
		},
		authManager.Login,
	)
	issuer := Settings.Auth.IssuerURL
	if issuer == "" {
		issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
	}
	p.openId, err = provider.New(
		goidc.ProfileOpenID,
		issuer,
		func(ctx context.Context) (keySet goidc.JSONWebKeySet, err error) {
			keySet = p.keySet
			return
		},
		provider.WithScopes(
			goidc.ScopeOpenID,
			goidc.ScopeProfile,
			goidc.ScopeEmail,
		),
		provider.WithGrantTypes(
			goidc.GrantClientCredentials,
			goidc.GrantAuthorizationCode,
			goidc.GrantRefreshToken,
		),
		provider.WithPKCERequired(goidc.CodeChallengeMethodSHA256),
		provider.WithTokenManager(tokenManager),
		provider.WithGrantManager(grantManager),
		provider.WithPolicies(authPolicy),
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	client := &goidc.Client{}
	client.ID = Settings.Auth.Client.ID
	client.Name = Settings.Auth.Client.Name
	client.Secret = Settings.Auth.Client.Secret
	client.TokenAuthnMethod = goidc.AuthnMethodSecretPost
	client.ScopeIDs = "openid profile email"
	client.GrantTypes = []goidc.GrantType{
		goidc.GrantClientCredentials,
		goidc.GrantAuthorizationCode,
		goidc.GrantRefreshToken,
	}
	client.ResponseTypes = []goidc.ResponseType{
		goidc.ResponseTypeCode,
	}
	client.RedirectURIs = []string{
		issuer + "/callback",
	}
	err = p.openId.SaveClient(context.Background(), client)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// NewAuthManager returns an authn manager.
func NewAuthManager(db *gorm.DB) (m *AuthManager) {
	m = &AuthManager{db: db}
	return
}

// AuthManager applies authN and AuthZ.
type AuthManager struct {
	db *gorm.DB
}

// Login provides the access token authentication.
func (r *AuthManager) Login(
	writer http.ResponseWriter,
	request *http.Request,
	session *goidc.AuthnSession) (status goidc.Status, err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	var userid, password string
	if session.Subject == "" {
		userid = request.PostFormValue("userid")
		password = request.PostFormValue("password")
	}
	if userid == "" || password == "" {
		err = r.renderPage(writer, request, session)
		status = goidc.StatusInProgress
		return
	}
	user := &model.User{}
	err = r.db.First(user, "name", userid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.renderPage(writer, request, session)
			status = goidc.StatusInProgress
			err = nil
		}
		return
	}
	err = secret.Decrypt(user)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if password != user.Password {
		err = r.renderPage(writer, request, session)
		status = goidc.StatusInProgress
		return
	}
	session.Subject = user.UUID
	status = goidc.StatusSuccess
	return
}

// renderPage renders the login page.
func (r *AuthManager) renderPage(writer http.ResponseWriter, _ *http.Request, session *goidc.AuthnSession) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	issuer := Settings.Auth.IssuerURL
	if issuer == "" {
		issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
	}
	// Simple login form HTML - POST to callback URL with session CallbackID
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Tackle Hub - Login</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        h1 { color: #333; }
        form { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        input { width: 100%; padding: 8px; margin: 10px 0; box-sizing: border-box; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; width: 100%; }
        button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <h1>Tackle Hub Login</h1>
    <form action="` + issuer + `/authorize/` + session.CallbackID + `" method="post">
        <div>
            <label>Username:</label>
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

// NewTokenManager returns a token manager.
func NewTokenManager(db *gorm.DB) (m *TokenManager) {
	m = &TokenManager{db: db}
	return
}

// TokenManager manages tokens.
type TokenManager struct {
	db *gorm.DB
}

// Save stores a token.
func (r *TokenManager) Save(ctx context.Context, token *goidc.Token) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{
		TokenId:    token.ID,
		GrantId:    token.GrantID,
		ClientId:   token.ClientID,
		Subject:    token.Subject,
		Type:       string(token.Type),
		Scopes:     token.Scopes,
		Issued:     asTime(token.CreatedAtTimestamp),
		Expiration: asTime(token.ExpiresAtTimestamp),
	}
	m.UserID, err = r.getUser(token)
	if err != nil {
		return
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.Save(m).Error
	return
}

// Token returns a token by id.
func (r *TokenManager) Token(ctx context.Context, id string) (token *goidc.Token, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.First(m, "tokenId", id).Error
	if err != nil {
		err = notFound(err)
		return
	}
	err = secret.Decrypt(m)
	if err != nil {
		return
	}
	token = &goidc.Token{
		ID:                 m.TokenId,
		GrantID:            m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		Type:               goidc.TokenType(m.Type),
		Scopes:             m.Scopes,
		CreatedAtTimestamp: asInt(m.Issued),
		ExpiresAtTimestamp: asInt(m.Expiration),
	}
	return
}

// ByRefreshToken returns a grant by refresh token.
func (r *TokenManager) ByRefreshToken(token string) (m *model.Token, err error) {
	err = r.db.First(m, "refreshToken", token).Error
	if err != nil {
		err = notFound(err)
		return
	}
	return
}

// Delete grant by id.
func (r *TokenManager) Delete(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.Delete(m, "tokenId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// DeleteByGrantID by grant by id.
func (r *TokenManager) DeleteByGrantID(_ context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.Delete(m, "grantId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// GetUser returns the user ID as appropriate.
func (r *TokenManager) getUser(token *goidc.Token) (userId *uint, err error) {
	grantManager := NewGrantManager(r.db)
	grant, err := grantManager.Grant(context.TODO(), token.GrantID)
	if err != nil {
		return
	}
	if grant.Type != goidc.GrantAuthorizationCode {
		return
	}
	user := &model.User{}
	err = r.db.First(user, "uuid", token.Subject).Error
	if err != nil {
		return
	}
	userId = &user.ID
	return
}

// NewKeyManager returns a configured key manager.
func NewKeyManager(db *gorm.DB) (m *KeyManager) {
	m = &KeyManager{db: db}
	return
}

// KeyManager manages RSA keys.
type KeyManager struct {
	db *gorm.DB
}

// KeySet returns a keyset.
// Rotation is applied.
func (r *KeyManager) KeySet() (keySet KeySet, err error) {
	var keyList []*model.RsaKey
	db := r.db.Order("id desc")
	err = db.Find(&keyList).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	created, err := r.rotate(keyList)
	if err != nil {
		return
	}
	keyList = append(created, keyList...)
	for _, m := range keyList {
		err = secret.Decrypt(m)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		b := []byte(m.PEM)
		decoded, _ := pem.Decode(b)
		var key *rsa.PrivateKey
		key, err = x509.ParsePKCS1PrivateKey(decoded.Bytes)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		jwKey := r.jwKey(m.ID, key)
		keySet.Keys = append(keySet.Keys, jwKey)
	}
	return
}

// rotate returns a new RSA key as determined
// by the rotation schedule.
func (r *KeyManager) rotate(keyList []*model.RsaKey) (created []*model.RsaKey, err error) {
	threshold := Settings.Auth.Key.Rotation
	for _, key := range keyList {
		age := time.Since(key.CreateTime)
		if age < threshold {
			return
		}
	}
	_, m := r.newKey()
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.Create(&m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	created = append(created, m)
	return
}

// newKey returns a new RSA key.
func (r *KeyManager) newKey() (key *rsa.PrivateKey, m *model.RsaKey) {
	m = &model.RsaKey{}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	b := x509.MarshalPKCS1PrivateKey(key)
	b = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: b,
	})
	m.PEM = string(b)
	return
}

// jwKey returns a goidc.JSONWebKey.
func (r *KeyManager) jwKey(id uint, k *rsa.PrivateKey) (k2 goidc.JSONWebKey) {
	k2.Key = strconv.Itoa(int(id))
	k2.Algorithm = "RS256"
	k2.Key = k
	return
}

func NewGrantManager(db *gorm.DB) (m *GrantManager) {
	m = &GrantManager{db: db}
	return
}

type GrantManager struct {
	db *gorm.DB
}

// Save stores a grant.
func (r *GrantManager) Save(_ context.Context, grant *goidc.Grant) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	if grant.Type == goidc.GrantAuthorizationCode {

	}
	m := &model.Grant{
		GrantId:      grant.ID,
		ClientId:     grant.ClientID,
		Subject:      grant.Subject,
		RefreshToken: grant.RefreshToken,
		AuthCode:     grant.AuthCode,
		Type:         string(grant.Type),
		Scopes:       grant.Scopes,
		Expiration:   asTime(grant.ExpiresAtTimestamp),
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.Save(m).Error
	return
}

// Grant returns a grant by id.
func (r *GrantManager) Grant(_ context.Context, id string) (grant *goidc.Grant, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.First(m, "grantId", id).Error
	if err != nil {
		err = notFound(err)
		return
	}
	grant, err = r.grant(m)

	return
}

// GrantByRefreshToken returns a grant by refresh token.
// Revocation is applied.
func (r *GrantManager) GrantByRefreshToken(_ context.Context, token string) (grant *goidc.Grant, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{
		RefreshToken: token,
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.First(m, "refreshToken", m.RefreshToken).Error
	if err != nil {
		err = notFound(err)
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
	grant, err = r.grant(m)
	return
}

// Delete a grant by id.
func (r *GrantManager) Delete(_ context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.Delete(m, "grantId", id).Error
	return
}

// DeleteByAuthCode delete a grant by auth code.
func (r *GrantManager) DeleteByAuthCode(_ context.Context, authCode string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.Delete(m, "authCode", authCode).Error
	return
}

// grant returns a goidc.Grant using
// the decrypted grant.
func (r *GrantManager) grant(m *model.Grant) (grant *goidc.Grant, err error) {
	err = secret.Decrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	grant = &goidc.Grant{
		ID:                 m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		RefreshToken:       m.RefreshToken,
		AuthCode:           m.AuthCode,
		Type:               goidc.GrantType(m.Type),
		Scopes:             m.Scopes,
		ExpiresAtTimestamp: asInt(m.Expiration),
	}
	return
}

// revoked enforces token revocation.
// When the access token associated with a granted by
// refresh token has been revoked, the grant expiration
// is updated to match the revocation timestamp.
func (r *GrantManager) revoked(grant *model.Grant) (err error) {
	tokenManager := NewTokenManager(r.db)
	token, err := tokenManager.ByRefreshToken(grant.RefreshToken)
	if err != nil {
		return
	}
	if !token.Revoked.IsZero() {
		grant.Expiration = token.Revoked
		err = r.db.Save(grant).Error
	}
	return
}

// orphaned imposes grant expiration when the
// user cannot be found.
func (r *GrantManager) orphaned(grant *model.Grant) (err error) {
	if grant.Type != string(goidc.GrantAuthorizationCode) {
		return
	}
	var count int64
	user := &model.User{}
	db := r.db.Model(user)
	db = db.Where("uuid", grant.Subject)
	err = r.db.Count(&count).Error
	if err != nil {
		return
	}
	if count == 0 {
		grant.Expiration = time.Now().UTC()
	}
	return
}

// notFound returns goidc.ErrNotFound when
// err IsA gorm.ErrRecordNotFound.
// Else, wrapped.
func notFound(err error) (e2 error) {
	if err == nil {
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		e2 = goidc.ErrNotFound
	} else {
		e2 = liberr.Wrap(err)
	}
	return
}

// asTime returns a time.Time for unix time.
func asTime(n int) (t time.Time) {
	t = time.Unix(int64(n), 0)
	t = t.UTC()
	return
}

// asInt returns unix time for time.Time.
func asInt(t time.Time) (i int) {
	t = t.UTC()
	i = int(t.Unix())
	return
}
