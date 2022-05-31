package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v10"
	"github.com/golang-jwt/jwt/v4"
)

type Provider interface {
	// Scopes decodes a list of scopes from the token.
	Scopes(token string) ([]Scope, error)
	// User parses preffered_username field from the token.
	User(token string) (user string, err error)
}

//
// Scope represents an authorization scope.
type Scope interface {
	// Allow determines whether the scope gives access to the resource with the method.
	Allow(resource string, method string) bool
}

//
// NoAuth provider always permits access.
type NoAuth struct{}

//
// Scopes decodes a list of scopes from the token.
// For the NoAuth provider, this just returns a single instance
// of the NoAuthScope.
func (r *NoAuth) Scopes(token string) (scopes []Scope, err error) {
	scopes = append(scopes, &NoAuthScope{})
	return
}

//
// User mocks username for NoAuth
func (r *NoAuth) User(token string) (name string, err error) {
	return "admin.noauth", nil
}

//
// NoAuthScope always permits access.
type NoAuthScope struct{}

//
// Check whether the scope gives access to the resource with the method.
func (r *NoAuthScope) Allow(_ string, _ string) (ok bool) {
	ok = true
	return
}

//
// NewKeycloak builds a new Keycloak auth provider.
func NewKeycloak(host, realm, id, secret string) (k Keycloak) {
	client := gocloak.NewClient(host)
	k = Keycloak{
		host:   host,
		realm:  realm,
		id:     id,
		secret: secret,
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
	id     string
	secret string
}

//
// Scopes decodes a list of scopes from the token.
func (r *Keycloak) Scopes(token string) (scopes []Scope, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	decoded, _, err := r.client.DecodeAccessToken(ctx, token, r.realm)
	if err != nil {
		err = errors.New("invalid token")
		return
	}
	if !decoded.Valid {
		err = errors.New("invalid token")
		return
	}
	claims, ok := decoded.Claims.(*jwt.MapClaims)
	if !ok || claims == nil {
		err = errors.New("invalid token")
		return
	}
	rawClaimScopes, ok := (*claims)["scope"].(string)
	if !ok {
		err = errors.New("invalid token")
		return
	}
	claimScopes := strings.Split(rawClaimScopes, " ")
	for _, s := range claimScopes {
		scope := r.newScope(s)
		scopes = append(scopes, &scope)
	}
	return
}

//
// NewKeycloakScope builds a Scope object from a string.
func (r *Keycloak) newScope(s string) (scope KeycloakScope) {
	if strings.Contains(s, ":") {
		segments := strings.Split(s, ":")
		scope.resource = segments[0]
		scope.method = segments[1]
	} else {
		scope.resource = s
	}
	return
}

//
// User resolves token to Keycloak username.
func (r *Keycloak) User(token string) (user string, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Token validity should be checked before in Scopes method
	_, claims, err := r.client.DecodeAccessToken(ctx, token, r.realm)

	// Get preferred_username from the token payload as the user
	user, ok := (*claims)["preferred_username"].(string)
	if !ok {
		err = errors.New("cannot parse preferred_username from token")
		return
	}
	return
}

//
// KeycloakScope is a scope decoded from a Keycloak token.
type KeycloakScope struct {
	resource string
	method   string
}

//
// Allow determines whether the scope gives access to the resource with the method.
func (r *KeycloakScope) Allow(resource string, method string) (ok bool) {
	ok = r.resource == resource && r.method == method
	return
}
