package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
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
	p.keySet = goidc.JSONWebKeySet{
		Keys: []goidc.JSONWebKey{
			{
				KeyID:     "kid-1",
				Algorithm: "RS256",
				// Key: yourPrivateKey,   // TODO: add actual key here
			},
		},
	}
	tokenManager := &TokenManager{
		db: db,
	}
	authManager := &AuthManager{
		db: db,
	}
	authPolicy := goidc.NewPolicy(
		"main",
		func(r *http.Request, client *goidc.Client, session *goidc.AuthnSession) bool {
			return true // apply to all requests for now
		},
		authManager.Login,
	)
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

type TokenManager struct {
	db *gorm.DB
}

func (r TokenManager) Save(ctx context.Context, token *goidc.Token) (err error) {
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
	err = r.db.First(user, "uuid", token.Subject).Error
	if err != nil {
		return
	}
	m.UserID = user.ID
	err = r.db.Save(m).Error
	return
}

func (r TokenManager) Token(ctx context.Context, id string) (token *goidc.Token, err error) {
	m := model.Token{}
	err = r.db.First(&m, "tokenId", id).Error
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

func (r TokenManager) Delete(ctx context.Context, id string) (err error) {
	m := model.Token{}
	err = r.db.Delete(m, "tokenId", id).Error
	return
}

func (r TokenManager) DeleteByGrantID(ctx context.Context, id string) (err error) {
	m := model.Token{}
	err = r.db.Delete(m, "grantId", id).Error
	return
}
