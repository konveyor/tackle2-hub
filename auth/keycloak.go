package auth

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
)

// NewKeycloak builds a new Keycloak auth provider.
func NewKeycloak(host, realm string) (p Provider) {
	client := gocloak.NewClient(host, gocloak.SetAuthRealms("auth/realms"), gocloak.SetAuthAdminRealms("auth/admin/realms"))
	client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	p = &Keycloak{
		host:   host,
		realm:  realm,
		client: client,
	}
	return
}

// Keycloak auth provider
type Keycloak struct {
	client *gocloak.GoCloak
	host   string
	realm  string
}

// NewToken creates a new signed token.
func (r Keycloak) NewToken(user string, scopes []string, claims jwt.MapClaims) (signed string, err error) {
	return
}

// Login and obtain a token.
func (r *Keycloak) Login(user, password string) (token Token, err error) {
	jwt, err := r.client.Login(
		context.TODO(),
		Settings.Auth.Keycloak.ClientID,
		Settings.Auth.Keycloak.ClientSecret,
		Settings.Auth.Keycloak.Realm,
		user,
		password,
	)
	if err == nil {
		token.Access = jwt.AccessToken
		token.Refresh = jwt.RefreshToken
		token.Expiry = jwt.ExpiresIn
	}
	return
}

// Refresh token.
func (r *Keycloak) Refresh(refresh string) (token Token, err error) {
	jwt, err := r.client.RefreshToken(
		context.TODO(),
		refresh,
		Settings.Auth.Keycloak.ClientID,
		Settings.Auth.Keycloak.ClientSecret,
		Settings.Auth.Keycloak.Realm,
	)
	if err == nil {
		token.Access = jwt.AccessToken
		token.Refresh = jwt.RefreshToken
		token.Expiry = jwt.ExpiresIn
	}
	return
}

// Authenticate the token
func (r *Keycloak) Authenticate(request *Request) (jwToken *jwt.Token, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	token := request.Token
	jwToken, _, err = r.client.DecodeAccessToken(ctx, token, r.realm)
	if err != nil || !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	claims, cast := jwToken.Claims.(jwt.MapClaims)
	if !cast {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	v, found := claims["preferred_username"]
	if !found {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	v, found = claims["scope"]
	if !found {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	return
}

// Scopes decodes a list of scopes from the token.
func (r *Keycloak) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	claims := jwToken.Claims.(jwt.MapClaims)
	for _, s := range strings.Fields(claims["scope"].(string)) {
		scope := BaseScope{}
		scope.With(s)
		scopes = append(scopes, &scope)
	}
	return
}

// User resolves token to Keycloak username.
func (r *Keycloak) User(jwToken *jwt.Token) (user string) {
	claims, _ := jwToken.Claims.(jwt.MapClaims)
	user, _ = claims["preferred_username"].(string)
	return
}
