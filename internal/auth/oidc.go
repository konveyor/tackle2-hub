package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"path"
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
		issuer = path.Join(
			Settings.Addon.Hub.URL,
			api.OIDCRoutes)
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
		path.Join(issuer, "/callback"),
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
	return
}

// NewTokenManager returns a token manager.
func NewTokenManager(db *gorm.DB) (m *TokenManager) {
	m = &TokenManager{db: db}
	return
}

type TokenManager struct {
	db *gorm.DB
}

func (r *TokenManager) Save(ctx context.Context, token *goidc.Token) (err error) {
	m := model.Token{
		TokenId:    token.ID,
		GrantId:    token.GrantID,
		ClientId:   token.ClientID,
		Subject:    token.Subject,
		Type:       string(token.Type),
		Scopes:     token.Scopes,
		Issued:     r.asTime(token.CreatedAtTimestamp),
		Expiration: r.asTime(token.ExpiresAtTimestamp),
	}
	user := &model.User{}
	err = r.db.First(user, "uuid", token.Subject).Error
	if err != nil {
		return
	}
	m.UserID = user.ID
	err = secret.Encrypt(m)
	if err != nil {
		return
	}
	err = r.db.Save(m).Error
	return
}

func (r *TokenManager) Token(ctx context.Context, id string) (token *goidc.Token, err error) {
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
	m := model.Token{}
	err = r.db.Delete(m, "tokenId", id).Error
	return
}

func (r *TokenManager) DeleteByGrantID(ctx context.Context, id string) (err error) {
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
