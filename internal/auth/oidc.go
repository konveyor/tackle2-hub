package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"gorm.io/gorm"
)

type BuiltinProvider struct {
	openId *provider.Provider
	keySet goidc.JSONWebKeySet
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
	p.keySet = goidc.JSONWebKeySet{
		Keys: []goidc.JSONWebKey{
			{
				KeyID:     "kid-1",
				Algorithm: "RS256",
				// Key: yourPrivateKey,   // TODO: add actual key here
			},
		},
	}
	authPolicy := goidc.NewPolicy(
		"main",
		func(r *http.Request, client *goidc.Client, session *goidc.AuthnSession) bool {
			return true // apply to all requests for now
		},
		func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (status goidc.Status, err error) {
			// TODO: Full authentication + authorization logic goes here:
			// 1. Check if this is local login (username/password)
			// 2. Or external IdP delegation (if idp=xxx parameter exists)
			// 3. Lookup user from DB
			// 4. Load roles → permissions → scopes
			// 5. Set as.Subject and as.Scopes accordingly
			user := &model.User{}
			as.Subject = user.UUID
			status = goidc.StatusSuccess
			return
		},
	)
	tokenManager := &TokenManager{
		db: db,
	}
	p.openId, err = provider.New(
		goidc.ProfileOpenID,
		Settings.Auth.Token.Key,
		func(ctx context.Context) (keySet goidc.JSONWebKeySet, err error) {
			keySet = p.keySet
			return
		},
		provider.WithGrantTypes(
			goidc.GrantAuthorizationCode,
			goidc.GrantRefreshToken,
		),
		provider.WithPKCERequired(goidc.CodeChallengeMethodSHA256),
		provider.WithPolicies(authPolicy),
		provider.WithTokenManager(tokenManager),
		// provider.WithClientManager(gormClientManager)
		// provider.WithAuthnSessionManager(gormSessionManager)
		// provider.WithGrantManager(gormGrantManager)
	)
	if err != nil {
		return
	}
	return
}

type TokenManager struct {
	db *gorm.DB
}

func (t TokenManager) Save(ctx context.Context, token *goidc.Token) (err error) {
	m := model.Token{
		TokenId:    token.ID,
		GrantId:    token.GrantID,
		ClientId:   token.ClientID,
		Subject:    token.Subject,
		Kind:       string(token.Type),
		Scopes:     token.Scopes,
		Issued:     token.CreatedAtTimestamp,
		Expiration: token.ExpiresAtTimestamp,
	}
	user := &model.User{}
	err = t.db.First(user, "uuid", token.Subject).Error
	if err != nil {
		return
	}
	m.UserID = user.ID
	err = t.db.Save(m).Error
	return
}

func (t TokenManager) Token(ctx context.Context, id string) (token *goidc.Token, err error) {
	m := model.Token{}
	err = t.db.First(&m, "tokenId", id).Error
	if err != nil {
		return
	}
	token = &goidc.Token{
		ID:                 m.TokenId,
		GrantID:            m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		Type:               goidc.TokenType(m.Kind),
		Scopes:             m.Scopes,
		CreatedAtTimestamp: m.Issued,
		ExpiresAtTimestamp: m.Expiration,
	}
	return
}

func (t TokenManager) Delete(ctx context.Context, id string) (err error) {
	m := model.Token{}
	err = t.db.Delete(m, "tokenId = ?", id).Error
	return
}

func (t TokenManager) DeleteByGrantID(ctx context.Context, id string) (err error) {
	m := model.Token{}
	err = t.db.Delete(m, "grantId = ?", id).Error
	return
}
