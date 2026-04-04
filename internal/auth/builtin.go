package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
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

// KeySet alias.
type KeySet = goidc.JSONWebKeySet

// Builtin
type Builtin struct {
	db       *gorm.DB
	openId   *provider.Provider
	keyCache KeyCache
	keySet   KeySet
}

// Handler returns an http handler.
func (p *Builtin) Handler() (h http.Handler) {
	h = p.openId.Handler()
	return
}

// UserKey returns a new key.
func (p *Builtin) UserKey(userId, password string, expiration time.Duration) (key APIKey, err error) {
	key, err = p.genKey()
	if err != nil {
		return
	}
	user := &model.User{}
	err = p.db.First(user, "UserId", userId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = &NotAuthenticated{
				Token: userId,
			}
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	err = secret.Decrypt(user)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if user.Password != password {
		err = &NotAuthenticated{
			Token: userId,
		}
		return
	}
	m := &model.APIKey{
		UserID:     &user.ID,
		Expiration: time.Now().Add(expiration),
		Secret:     key.Secret,
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = p.db.Create(m).Error
	return
}

// TaskKey returns a new key.
func (p *Builtin) TaskKey(taskId uint, expiration time.Duration) (key APIKey, err error) {
	key, err = p.genKey()
	if err != nil {
		return
	}
	owner := &model.User{}
	err = p.db.First(owner, taskId).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	m := &model.APIKey{
		TaskID:     &owner.ID,
		Expiration: time.Now().Add(expiration),
		Secret:     key.Secret,
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = p.db.Create(m).Error
	return
}

// Authenticate a web request.
func (p *Builtin) Authenticate(request *Request) (jwToken *jwt.Token, err error) {
	defer func() {
		if errors.Is(err, &NotValid{}) {
			Log.V(2).Info("[builtin] " + err.Error())
		}
	}()
	bearer, err := p.extractBearer(request)
	if err != nil {
		return
	}
	jwToken, err = jwt.Parse(
		bearer,
		func(jwToken *jwt.Token) (key any, err error) {
			_, cast := jwToken.Method.(*jwt.SigningMethodRSA)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{Token: bearer})
				return
			}
			kid, found := jwToken.Header["kid"]
			if !found {
				err = liberr.Wrap(&NotAuthenticated{Token: bearer})
				return
			}
			key, err = p.keySet.Key(kid.(string))
			return
		})
	if err != nil {
		jwToken, err = jwt.Parse(
			bearer,
			func(jwToken *jwt.Token) (secret any, err error) {
				_, cast := jwToken.Method.(*jwt.SigningMethodHMAC)
				if !cast {
					err = liberr.Wrap(&NotAuthenticated{Token: bearer})
					return
				}
				secret = []byte(Settings.Auth.Token.Key)
				return
			})
	}
	if err == nil {
		err = p.validateToken(jwToken)
		return
	}
	key, err := p.keyCache.Get(bearer)
	if err == nil {
		token := jwt.New(jwt.SigningMethodHS512)
		jwtClaims := token.Claims.(jwt.MapClaims)
		jwtClaims["scopes"] = strings.Join(key.Scopes, " ")
		jwtClaims["subject"] = key.User
	}
	return
}

// Revoke an access token.
func (p *Builtin) Revoke(token *jwt.Token) (err error) {
	return
}

func (r *Builtin) User(jwToken *jwt.Token) (user string) {
	claims := jwToken.Claims.(jwt.MapClaims)
	user = claims["user"].(string)
	return
}

// Scopes returns a list of scopes.
func (p *Builtin) Scopes(jwToken *jwt.Token) (scopes []Scope) {
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

// extractBearer returns the token
func (p *Builtin) extractBearer(request *Request) (bearer string, err error) {
	splitToken := strings.Fields(request.Token)
	if len(splitToken) != 2 || strings.ToLower(splitToken[0]) != "bearer" {
		err = liberr.Wrap(&NotValid{Token: request.Token})
		return
	}
	bearer = splitToken[1]
	return
}

// validateToken determines if the token is valid..
func (p *Builtin) validateToken(jwToken *jwt.Token) (err error) {
	if !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{Token: jwToken.Raw})
		return
	}
	claims, cast := jwToken.Claims.(jwt.MapClaims)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Claims not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	v, found := claims["sub"]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not string.",
				Token:  jwToken.Raw,
			})
		return
	}
	v, found = claims["scope"]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not string.",
				Token:  jwToken.Raw,
			})
		return
	}
	return
}

// genKey returns a new generated key.
func (p *Builtin) genKey() (key APIKey, err error) {
	prefix := "apikey_"
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	generated := base64.RawURLEncoding.EncodeToString(b)
	key.Secret = prefix + generated
	return
}

// NewBuiltin returns a configured provider.
func NewBuiltin(db *gorm.DB) (builtin *Builtin, err error) {
	builtin = &Builtin{
		keyCache: KeyCache{db: db},
		db:       db,
	}
	grantManager := NewGrantManager(db)
	keyManager := NewKeyManager(db)
	authManager := NewAuthManager(db)
	tokenManager := NewTokenManager(db)
	builtin.keySet, err = keyManager.KeySet()
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
	builtin.openId, err = provider.New(
		goidc.ProfileOpenID,
		issuer,
		func(ctx context.Context) (keySet goidc.JSONWebKeySet, err error) {
			keySet = builtin.keySet
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
		provider.WithTokenOptions(grantManager.tokenOptions),
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
	err = builtin.openId.SaveClient(context.Background(), client)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
