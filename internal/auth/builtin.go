package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
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
	keyCache *KeyCache
	keySet   KeySet
}

// Handler returns an http handler.
func (p *Builtin) Handler() (h http.Handler) {
	h = p.openId.Handler()
	return
}

// Grant the key request.
func (p *Builtin) Grant(kr KeyRequest) (key APIKey, err error) {
	key, err = p.genKey(kr.Lifespan)
	if err != nil {
		return
	}
	m := &model.APIKey{
		Expiration: key.Expiration,
		Digest:     key.Digest,
	}
	if kr.TaskID > 0 {
		task := &model.Task{}
		err = p.db.First(task, kr.TaskID).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		m.TaskID = &task.ID
	}
	if kr.Userid != "" {
		user := &model.User{}
		err = p.db.First(user, "Userid", kr.Userid).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = &NotAuthenticated{
					Token: kr.Userid,
				}
			} else {
				err = liberr.Wrap(err)
			}
			return
		}
		if !secret.MatchPassword(kr.Password, user.Password) {
			err = &NotAuthenticated{
				Token: kr.Userid,
			}
			return
		}
		m.UserID = &user.ID
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
			jwk, findErr := p.keySet.Key(kid.(string))
			if findErr != nil {
				err = liberr.Wrap(findErr)
				return
			}
			privateKey, cast := jwk.Key.(*rsa.PrivateKey)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{Token: bearer})
				return
			}
			key = &privateKey.PublicKey
			return
		},
		jwt.WithoutClaimsValidation())
	if err != nil {
		jwToken, err = jwt.Parse(
			bearer,
			func(jwToken *jwt.Token) (secret any, err error) {
				_, cast := jwToken.Method.(*jwt.SigningMethodHMAC)
				if !cast {
					err = liberr.Wrap(&NotAuthenticated{Token: bearer})
					return
				}
				secret = []byte(Settings.Token.Key)
				return
			},
			jwt.WithoutClaimsValidation())
	}
	if err == nil {
		err = p.validToken(jwToken)
		return
	}
	key, err := p.keyCache.Get(bearer)
	if err == nil {
		jwToken = jwt.New(jwt.SigningMethodHS512)
		jwtClaims := jwToken.Claims.(jwt.MapClaims)
		jwtClaims[ClaimScope] = strings.Join(key.Scopes, " ")
		jwtClaims[ClaimSub] = key.User
		return
	}
	err = liberr.Wrap(&NotAuthenticated{Token: bearer})
	return
}

// Revoke an access token.
func (p *Builtin) Revoke(token *jwt.Token) (err error) {
	return
}

// Delete an api key
func (p *Builtin) Delete(digest string) (err error) {
	p.keyCache.Delete(digest)
	m := &model.APIKey{}
	err = p.db.Delete(m, "digest", digest).Error
	return
}

func (r *Builtin) User(jwToken *jwt.Token) (user string) {
	claims := jwToken.Claims.(jwt.MapClaims)
	v := claims[ClaimSub]
	if s, cast := v.(string); cast {
		user = s
	}
	return
}

// Scopes returns a list of scopes.
func (p *Builtin) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	claims := jwToken.Claims.(jwt.MapClaims)
	v := claims[ClaimScope]
	if sList, cast := v.(string); cast {
		for _, s := range strings.Fields(sList) {
			scope := &BaseScope{}
			scope.With(s)
			scopes = append(
				scopes,
				scope)
		}
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

// validToken returns an error if not valid.
func (p *Builtin) validToken(jwToken *jwt.Token) (err error) {
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
	v, found := claims[ClaimSub]
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
	v, found = claims[ClaimScope]
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
	v, found = claims[ClaimExp]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Exp not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	f64, cast := v.(float64)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Exp not float64.",
				Token:  jwToken.Raw,
			})
		return
	}
	expiration := time.Unix(int64(f64), 0)
	if expiration.Before(time.Now()) {
		err = &NotValid{
			Reason: "Token expired.",
			Token:  jwToken.Raw,
		}
		return
	}
	return
}

// genKey returns a new generated key.
func (p *Builtin) genKey(lifespan time.Duration) (key APIKey, err error) {
	prefix := "apikey_"
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	generated := base64.RawURLEncoding.EncodeToString(b)
	key.Secret = prefix + generated
	key.Digest = secret.Hash(key.Secret)
	key.Expiration = time.Now().Add(lifespan)
	return
}

// NewBuiltin returns a configured provider.
func NewBuiltin(db *gorm.DB) (builtin *Builtin, err error) {
	builtin = &Builtin{
		keyCache: NewCache(db),
		db:       db,
	}
	//
	// Managers
	grantManager := NewGrantManager(db)
	keyManager := NewKeyManager(db)
	authManager := NewAuthManager(db)
	tokenManager := NewTokenManager(db)
	builtin.keySet, err = keyManager.KeySet()
	if err != nil {
		return
	}
	//
	// Auth policy
	authPolicy := goidc.NewPolicy(
		"main",
		func(*http.Request, *goidc.Client, *goidc.AuthnSession) bool {
			return true
		},
		authManager.Login,
	)
	issuer := Settings.IssuerURL
	if issuer == "" {
		issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
	}
	tokenOptions := func(
		_ context.Context,
		_ *goidc.Grant,
		_ *goidc.Client) (options goidc.TokenOptions) {
		options = goidc.NewJWTTokenOptions(goidc.RS256, Settings.Token.Lifespan)
		return
	}
	//
	// userInfoClaims returns user profile claims for the /userinfo endpoint.
	userInfoClaims := func(
		_ context.Context,
		grant *goidc.Grant) (claims map[string]any) {
		claims = make(map[string]any)
		user := &model.User{}
		err := db.First(user, "Subject", grant.Subject).Error
		if err != nil {
			Log.Error(err, "")
			return
		}
		claims[goidc.ClaimSubject] = user.Subject
		claims[goidc.ClaimPreferredUsername] = user.Userid
		if user.Email != "" {
			claims[goidc.ClaimEmail] = user.Email
			claims[goidc.ClaimEmailVerified] = true
		}
		return
	}
	//
	// Provider
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
		provider.WithRefreshTokenLifetime(Settings.Token.RefreshLifespan),
		provider.WithTokenOptions(tokenOptions),
		provider.WithTokenManager(tokenManager),
		provider.WithGrantManager(grantManager),
		provider.WithPolicies(authPolicy),
		provider.WithUserInfoClaims(userInfoClaims),
		provider.WithIDTokenSignatureAlgs(goidc.RS256),
		provider.WithUserInfoSignatureAlgs(goidc.RS256),
		provider.WithResourceIndicators(issuer),
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	//
	// Client
	redirectURIs := Settings.Auth.Client.RedirectURIs
	if len(redirectURIs) == 0 {
		// When not explicitly configured, allow callback to issuer for
		// development/testing. Production deployments MUST set
		// OIDC_CLIENT_REDIRECT_URIS to the UI callback URL.
		redirectURIs = []string{
			issuer + "/callback",
		}
	}
	client := &goidc.Client{}
	client.ID = Settings.Auth.Client.ID
	client.IsPublic()
	client.Name = Settings.Auth.Client.Name
	client.ScopeIDs = "openid profile email"
	client.RedirectURIs = redirectURIs
	client.TokenAuthnMethod = goidc.AuthnMethodNone
	client.GrantTypes = []goidc.GrantType{
		goidc.GrantClientCredentials,
		goidc.GrantAuthorizationCode,
		goidc.GrantRefreshToken,
	}
	client.ResponseTypes = []goidc.ResponseType{
		goidc.ResponseTypeCode,
	}
	if Settings.Auth.Client.Secret != "" {
		client.Secret = Settings.Auth.Client.Secret
		client.TokenAuthnMethod = goidc.AuthnMethodSecretPost
	}
	err = builtin.openId.SaveClient(context.Background(), client)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
