package auth

import (
	"context"
	"crypto/tls"
	"github.com/Nerzal/gocloak/v10"
	"github.com/golang-jwt/jwt/v4"
	liberr "github.com/konveyor/controller/pkg/error"
	"strings"
	"time"
)

//
// NewKeycloak builds a new Keycloak auth provider.
func NewKeycloak(host, realm string) (p Provider) {
	client := gocloak.NewClient(host)
	client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	p = &Keycloak{
		host:   host,
		realm:  realm,
		client: client,
	}
	return
}

//
// Keycloak auth provider
type Keycloak struct {
	client gocloak.GoCloak
	host   string
	realm  string
}

//
// NewToken creates a new signed token.
func (r Keycloak) NewToken(user string, scopes []string, claims jwt.MapClaims) (signed string, err error) {
	return
}

//
// Authenticate the token
func (r *Keycloak) Authenticate(token string) (jwToken *jwt.Token, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	jwToken, _, err = r.client.DecodeAccessToken(ctx, token, r.realm)
	if err != nil || !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	claims, cast := jwToken.Claims.(*jwt.MapClaims)
	if !cast {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	v, found := (*claims)["preferred_username"]
	if !found {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	v, found = (*claims)["scope"]
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

//
// Scopes decodes a list of scopes from the token.
func (r *Keycloak) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	claims := jwToken.Claims.(*jwt.MapClaims)
	for _, s := range strings.Fields((*claims)["scope"].(string)) {
		scope := BaseScope{}
		scope.With(s)
		scopes = append(scopes, &scope)
	}
	return
}

//
// User resolves token to Keycloak username.
func (r *Keycloak) User(jwToken *jwt.Token) (user string) {
	claims, _ := jwToken.Claims.(*jwt.MapClaims)
	user, _ = (*claims)["preferred_username"].(string)
	return
}
