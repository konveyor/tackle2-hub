package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/Nerzal/gocloak/v10"
	"github.com/golang-jwt/jwt/v4"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"strings"
	"time"
)

var log = logging.WithName("auth")

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
	name = "admin.noauth"
	return
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
func NewKeycloak(host, realm, id, secret, admin, pass, adminRealm string) (k Keycloak) {
	client := gocloak.NewClient(host)
	k = Keycloak{
		host:       host,
		realm:      realm,
		id:         id,
		secret:     secret,
		client:     client,
		admin:      admin,
		pass:       pass,
		adminRealm: adminRealm,
	}
	return
}

//
// Keycloak auth provider
type Keycloak struct {
	client     gocloak.GoCloak
	host       string
	realm      string
	id         string
	secret     string
	admin      string
	pass       string
	adminRealm string
}

//
// EnsureRoles ensures that hub roles and scopes are present in keycloak by
// creating them if they are missing and assigning scope mappings.
func (r *Keycloak) EnsureRoles() (err error) {
	hubRoles, err := LoadRoles(Settings.Auth.RolePath)
	if err != nil {
		return
	}
	token, err := r.login()
	if err != nil {
		return
	}
	scopeMap, err := r.scopeMap(token)
	if err != nil {
		return
	}
	roleMap, err := r.realmRoleMap(token)
	if err != nil {
		return
	}

	// create missing roles and scopes, and build mapping of scopes to roles
	scopesToRoles := make(map[string][]gocloak.Role)
	for _, role := range hubRoles {
		if _, ok := roleMap[role.Name]; !ok {
			realmRole := gocloak.Role{Name: &role.Name}
			id, kErr := r.client.CreateRealmRole(context.Background(), token.AccessToken, r.realm, realmRole)
			if kErr != nil {
				err = liberr.Wrap(kErr)
				return
			}
			realmRole.ID = &id
			roleMap[role.Name] = realmRole
			log.Info("Created realm role.", "role", role.Name, "realm", r.realm)
		}

		for _, res := range role.Resources {
			for _, verb := range res.Verbs {
				scopeName := fmt.Sprintf("%s:%s", res.Name, verb)
				if _, ok := scopeMap[scopeName]; !ok {
					scope := gocloak.ClientScope{Name: &scopeName}
					id, kErr := r.client.CreateClientScope(context.Background(), token.AccessToken, r.realm, scope)
					if kErr != nil {
						err = liberr.Wrap(kErr)
						return
					}
					scope.ID = &id
					scopeMap[scopeName] = scope
					log.Info("Created client scope.", "scope", scopeName, "realm", r.realm)
				}
				scopesToRoles[*scopeMap[scopeName].ID] = append(scopesToRoles[*scopeMap[scopeName].ID], roleMap[role.Name])
			}
		}

		// assign roles to users
		realmRole := roleMap[role.Name]
		for _, username := range role.Users {
			err = r.addRoleToUser(token, username, realmRole)
			if err != nil {
				return
			}
		}
	}

	// create scope mappings and add default scopes to client
	client, err := r.hubClient(token)
	if err != nil {
		return
	}
	for sid, roles := range scopesToRoles {
		err = r.client.CreateClientScopesScopeMappingsRealmRoles(context.Background(), token.AccessToken, r.realm, sid, roles)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = r.client.AddDefaultScopeToClient(context.Background(), token.AccessToken, r.realm, *client.ID, sid)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	return
}

//
// hubClient returns the hub client.
func (r *Keycloak) hubClient(token *gocloak.JWT) (client gocloak.Client, err error) {
	max := 1
	clientParams := gocloak.GetClientsParams{ClientID: &r.id, Max: &max}
	clients, err := r.client.GetClients(context.Background(), token.AccessToken, r.realm, clientParams)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if len(clients) == 0 {
		err = liberr.New("Could not find client.", "realm", r.realm, "client", r.id)
		return
	}
	client = *clients[0]
	return
}

//
// addRoleToUser adds a role to a user.
func (r *Keycloak) addRoleToUser(token *gocloak.JWT, username string, role gocloak.Role) (err error) {
	exact := true
	userParams := gocloak.GetUsersParams{Username: &username, Exact: &exact}
	users, err := r.client.GetUsers(context.Background(), token.AccessToken, r.realm, userParams)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if len(users) != 1 {
		log.Info("Couldn't find user to add role.", "user", username, "role", role.Name, "realm", r.realm)
		return
	}

	roles := []gocloak.Role{role}
	err = r.client.AddRealmRoleToUser(context.Background(), token.AccessToken, r.realm, *users[0].ID, roles)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// realmRoleMap generates a mapping of realm role names to ids
func (r *Keycloak) realmRoleMap(token *gocloak.JWT) (roleMap map[string]gocloak.Role, err error) {
	roleMap = make(map[string]gocloak.Role)
	roleParams := gocloak.GetRoleParams{}
	realmRoles, err := r.client.GetRealmRoles(context.Background(), token.AccessToken, r.realm, roleParams)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	for _, role := range realmRoles {
		roleMap[*role.Name] = *role
	}

	return
}

//
// scopeMap generates a mapping of client scope names to ids
func (r *Keycloak) scopeMap(token *gocloak.JWT) (scopeMap map[string]gocloak.ClientScope, err error) {
	scopeMap = make(map[string]gocloak.ClientScope)
	clientScopes, err := r.client.GetClientScopes(context.Background(), token.AccessToken, r.realm)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	for _, scope := range clientScopes {
		scopeMap[*scope.Name] = *scope
	}

	return
}

//
// login logs into the keycloak admin-cli client as the administrator.
func (r *Keycloak) login() (token *gocloak.JWT, err error) {
	token, err = r.client.LoginAdmin(context.Background(), r.admin, r.pass, r.adminRealm)
	if err != nil {
		return
	}
	return
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
	_, claims, err := r.client.DecodeAccessToken(ctx, token, r.realm)
	if err != nil {
		return
	}
	user, found := (*claims)["preferred_username"].(string)
	if !found {
		err = errors.New("preferred_username not found in token")
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
