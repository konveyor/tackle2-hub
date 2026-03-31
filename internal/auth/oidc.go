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
	"sync"
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

var OIDC *BuiltinProvider

type KeySet = goidc.JSONWebKeySet

type BuiltinProvider struct {
	openId *provider.Provider
	keySet KeySet
}

func (p *BuiltinProvider) Handler() http.Handler {
	return p.openId.Handler()
}

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

// New creates a new OIDC Provider for the Hub
func New(db *gorm.DB) (p *BuiltinProvider, err error) {
	p = &BuiltinProvider{}
	grantManager := NewGrantManager()
	keyManager := NewKeyManager(db)
	authManager := NewAuthManager(db)
	tokenManager := NewTokenManager(db, grantManager)
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
	client.RedirectURIs = []string{
		issuer + "/callback",
	}
	err = p.openId.SaveClient(context.Background(), client)
	if err != nil {
		return
	}
	return
}

// NewAuthManager returns an authn manager.
func NewAuthManager(db *gorm.DB) (m *AuthManager) {
	m = &AuthManager{db: db}
	return
}

type AuthManager struct {
	db *gorm.DB
}

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
		err = r.renderPage(writer, request)
		status = goidc.StatusInProgress
		return
	}
	user := &model.User{}
	err = r.db.First(user, "name", userid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.renderPage(writer, request)
			status = goidc.StatusInProgress
			err = nil
		}
		return
	}
	err = secret.Decrypt(user)
	if err != nil {
		return
	}
	if password != user.Password {
		err = r.renderPage(writer, request)
		status = goidc.StatusInProgress
		return
	}
	session.Subject = user.UUID
	status = goidc.StatusSuccess
	return
}

func (r *AuthManager) renderPage(writer http.ResponseWriter, request *http.Request) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	// Simple login form HTML
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
    <form method="post">
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
	return
}

// NewTokenManager returns a token manager.
func NewTokenManager(db *gorm.DB, grantManager *GrantManager) (m *TokenManager) {
	m = &TokenManager{
		grantManager: grantManager,
		db:           db,
	}
	return
}

type TokenManager struct {
	grantManager *GrantManager
	db           *gorm.DB
}

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
		Issued:     r.asTime(token.CreatedAtTimestamp),
		Expiration: r.asTime(token.ExpiresAtTimestamp),
	}
	grant, err := r.grantManager.Grant(ctx, token.GrantID)
	if err != nil {
		return
	}
	if grant.Type != goidc.GrantClientCredentials {
		user := &model.User{}
		err = r.db.First(user, "uuid", token.Subject).Error
		if err != nil {
			return
		}
		m.UserID = &user.ID
	}
	err = secret.Encrypt(m)
	if err != nil {
		return
	}
	err = r.db.Save(m).Error
	return
}

func (r *TokenManager) Token(ctx context.Context, id string) (token *goidc.Token, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := model.Token{}
	err = r.db.First(&m, "tokenId", id).Error
	if err != nil {
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
		CreatedAtTimestamp: r.asInt(m.Issued),
		ExpiresAtTimestamp: r.asInt(m.Expiration),
	}
	return
}

func (r *TokenManager) Delete(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := model.Token{}
	err = r.db.Delete(m, "tokenId", id).Error
	return
}

func (r *TokenManager) DeleteByGrantID(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := model.Token{}
	err = r.db.Delete(m, "grantId", id).Error
	return
}

func (r *TokenManager) asTime(n int) (t time.Time) {
	t = time.Unix(int64(n), 0)
	t = t.UTC()
	return
}

func (r *TokenManager) asInt(t time.Time) (i int) {
	t = t.UTC()
	i = int(t.Unix())
	return
}

func NewKeyManager(db *gorm.DB) (m *KeyManager) {
	m = &KeyManager{db: db}
	return
}

type KeyManager struct {
	db *gorm.DB
}

func (r *KeyManager) KeySet() (keySet KeySet, err error) {
	var keyList []*model.RsaKey
	db := r.db.Order("id desc")
	err = db.Find(&keyList).Error
	if err != nil {
		return
	}
	for _, m := range keyList {
		err = secret.Decrypt(m)
		if err != nil {
			return
		}
		b := []byte(m.PEM)
		decoded, _ := pem.Decode(b)
		var key *rsa.PrivateKey
		key, err = x509.ParsePKCS1PrivateKey(decoded.Bytes)
		if err != nil {
			return
		}
		jwKey := r.jwKey(m.ID, key)
		keySet.Keys = append(keySet.Keys, jwKey)
	}
	if len(keySet.Keys) == 0 {
		key, m := r.newKey()
		jwKey := r.jwKey(m.ID, key)
		keySet.Keys = append(keySet.Keys, jwKey)
		err = secret.Encrypt(m)
		if err != nil {
			return
		}
		err = db.Create(&m).Error
		if err != nil {
			return
		}
	}
	return
}

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

func (r *KeyManager) jwKey(id uint, k *rsa.PrivateKey) (k2 goidc.JSONWebKey) {
	k2.Key = strconv.Itoa(int(id))
	k2.Algorithm = "RS256"
	k2.Key = k
	return
}

func NewGrantManager() (m *GrantManager) {
	m = &GrantManager{
		byId:           make(map[string]*goidc.Grant),
		byRefreshToken: make(map[string]string),
		byAuthCode:     make(map[string]string),
	}
	return
}

type GrantManager struct {
	mutex          sync.RWMutex
	byId           map[string]*goidc.Grant
	byRefreshToken map[string]string
	byAuthCode     map[string]string
}

func (g *GrantManager) Save(ctx context.Context, grant *goidc.Grant) (err error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.byId[grant.ID] = grant
	g.byRefreshToken[grant.RefreshToken] = grant.ID
	g.byAuthCode[grant.AuthCode] = grant.ID
	return
}

func (g *GrantManager) Grant(ctx context.Context, id string) (grant *goidc.Grant, err error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	grant, found := g.byId[id]
	if !found {
		err = goidc.ErrNotFound
	}
	return
}

func (g *GrantManager) GrantByRefreshToken(ctx context.Context, token string) (grant *goidc.Grant, err error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	id := g.byRefreshToken[token]
	grant, found := g.byId[id]
	if !found {
		err = goidc.ErrNotFound
	}
	return
}

func (g *GrantManager) Delete(ctx context.Context, id string) (err error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.delete(id)
	return
}

func (g *GrantManager) DeleteByAuthCode(ctx context.Context, code string) (err error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	id, found := g.byAuthCode[code]
	if !found {
		return
	}
	g.delete(id)
	return
}

func (g *GrantManager) delete(id string) {
	grant, found := g.byId[id]
	if !found {
		return
	}
	delete(g.byId, id)
	delete(g.byRefreshToken, grant.RefreshToken)
	delete(g.byAuthCode, grant.AuthCode)
	return
}
