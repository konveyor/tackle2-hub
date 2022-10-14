package auth

import (
	"context"
	"crypto/tls"
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
	client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
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
	token      *gocloak.JWT
}

//
// Realm is a container for the users,
// scopes, and roles that exist in the
// hub's keycloak realm.
type Realm struct {
	Users  map[string]gocloak.User
	Scopes map[string]gocloak.ClientScope
	Roles  map[string]gocloak.Role
}

//
// Reconcile ensures that the Hub realm
// exists and the expected clients, roles, scopes,
// and users are present in it.
func (r *Keycloak) Reconcile() (err error) {
	err = r.login()
	if err != nil {
		return
	}

	err = r.ensureRealm()
	if err != nil {
		return
	}

	err = r.ensureClient()
	if err != nil {
		return
	}

	realm, err := r.loadRealm()
	if err != nil {
		return
	}

	err = r.ensureRoles(realm)
	if err != nil {
		return
	}

	err = r.ensureUsers(realm)
	if err != nil {
		return
	}

	log.Info("Realm synced.")

	return
}

//
// loadRealm loads all the scopes, roles, and users from
// the hub keycloak realm and populates a Realm struct.
func (r *Keycloak) loadRealm() (realm *Realm, err error) {
	realm = &Realm{}
	realm.Scopes, err = r.scopeMap()
	if err != nil {
		return
	}
	realm.Roles, err = r.realmRoleMap()
	if err != nil {
		return
	}
	realm.Users, err = r.userMap()
	if err != nil {
		return
	}

	return
}

//
// ensureRealm ensures that the hub realm exists.
func (r *Keycloak) ensureRealm() (err error) {
	_, err = r.client.GetRealm(context.Background(), r.token.AccessToken, r.realm)
	if err != nil {
		if strings.HasPrefix(err.Error(), "404") {
			displayName := fmt.Sprintf("%s realm", r.realm)
			enabled := true
			realm := gocloak.RealmRepresentation{
				Realm:       &r.realm,
				DisplayName: &displayName,
				Enabled:     &enabled,
			}
			log.Info("Creating realm.", "realm", r.realm)
			_, err = r.client.CreateRealm(context.Background(), r.token.AccessToken, realm)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			return
		}
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// ensureClient ensures that the hub client exists.
func (r *Keycloak) ensureClient() (err error) {
	var found bool
	_, found, err = r.getClient(r.id)
	if err != nil {
		return
	}
	if found {
		return
	}

	enabled := true
	newClient := gocloak.Client{
		ClientID:                  &r.id,
		StandardFlowEnabled:       &enabled,
		DirectAccessGrantsEnabled: &enabled,
		PublicClient:              &enabled,
		RedirectURIs:              &[]string{"*"},
		WebOrigins:                &[]string{"*"},
	}
	log.Info("Creating client.", "client", r.id)
	_, err = r.client.CreateClient(context.Background(), r.token.AccessToken, r.realm, newClient)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// ensureUsers ensures that the hub users exist and have the necessary roles.
func (r *Keycloak) ensureUsers(realm *Realm) (err error) {
	users, err := LoadUsers(Settings.Auth.UserPath)
	if err != nil {
		return
	}

	var allRoles []gocloak.Role
	for _, role := range realm.Roles {
		allRoles = append(allRoles, role)
	}

	for _, user := range users {
		if _, found := realm.Users[user.Name]; !found {
			enabled := true
			u := gocloak.User{
				Username: &user.Name,
				Enabled:  &enabled,
			}
			if Settings.Keycloak.RequirePasswordUpdate {
				u.RequiredActions = &[]string{"UPDATE_PASSWORD"}
			}
			log.Info("Creating user.", "user", user.Name)
			userid, kErr := r.client.CreateUser(context.Background(), r.token.AccessToken, r.realm, u)
			if kErr != nil {
				err = liberr.Wrap(kErr)
				return
			}
			u.ID = &userid
			err = r.client.SetPassword(
				context.Background(),
				r.token.AccessToken,
				userid,
				r.realm,
				user.Password,
				Settings.Keycloak.RequirePasswordUpdate,
			)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			realm.Users[user.Name] = u
		} else {
			log.Info("Removing any existing roles from user.", "user", user.Name)
			err = r.client.DeleteRealmRoleFromUser(
				context.Background(), r.token.AccessToken, r.realm, *realm.Users[user.Name].ID, allRoles,
			)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}

		var realmRoles []gocloak.Role
		for _, role := range user.Roles {
			realmRole, found := realm.Roles[role]
			if !found {
				continue
			}
			realmRoles = append(realmRoles, realmRole)
		}
		log.Info("Applying roles to user.", "user", user.Name, "roles", user.Roles)
		err = r.client.AddRealmRoleToUser(
			context.Background(), r.token.AccessToken, r.realm, *realm.Users[user.Name].ID, realmRoles,
		)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	return
}

//
// ensureRoles ensures that hub roles and scopes are present in keycloak by
// creating them if they are missing and assigning scope mappings.
func (r *Keycloak) ensureRoles(realm *Realm) (err error) {
	hubRoles, err := LoadRoles(Settings.Auth.RolePath)
	if err != nil {
		return
	}

	// create missing roles and scopes, and build mapping of scopes to roles
	scopesToRoles := make(map[string][]gocloak.Role)
	for i := range hubRoles {
		role := hubRoles[i]
		if _, found := realm.Roles[role.Name]; !found {
			_, kErr := r.client.CreateRealmRole(
				context.Background(), r.token.AccessToken, r.realm, gocloak.Role{Name: &role.Name},
			)
			if kErr != nil {
				err = liberr.Wrap(kErr)
				return
			}
			realmRole, kErr := r.client.GetRealmRole(
				context.Background(), r.token.AccessToken, r.realm, role.Name,
			)
			if kErr != nil {
				err = liberr.Wrap(kErr)
				return
			}
			realm.Roles[role.Name] = *realmRole
			log.Info("Created realm role.", "role", role.Name, "realm", r.realm)
		}

		for _, res := range role.Resources {
			for _, verb := range res.Verbs {
				scopeName := fmt.Sprintf("%s:%s", res.Name, verb)
				protocol := "openid-connect"
				if _, found := realm.Scopes[scopeName]; !found {
					scope := gocloak.ClientScope{Name: &scopeName, Protocol: &protocol}
					id, kErr := r.client.CreateClientScope(
						context.Background(), r.token.AccessToken, r.realm, scope,
					)
					if kErr != nil {
						err = liberr.Wrap(kErr)
						return
					}
					scope.ID = &id
					realm.Scopes[scopeName] = scope
					log.Info("Created client scope.", "scope", scopeName, "realm", r.realm)
				}
				scopesToRoles[*realm.Scopes[scopeName].ID] = append(
					scopesToRoles[*realm.Scopes[scopeName].ID], realm.Roles[role.Name],
				)
			}
		}
	}

	c, found, err := r.getClient(r.id)
	if err != nil {
		return
	}
	if !found {
		err = liberr.New("Could not find client.", "realm", r.realm, "client", r.id)
		return
	}
	for sid, roles := range scopesToRoles {
		log.Info("Syncing scope mappings.", "scope", sid, "roles", roles)
		// get the roles that are already mapped to this client scope
		var existingRoles []*gocloak.Role
		existingRoles, err = r.client.GetClientScopesScopeMappingsRealmRoles(
			context.Background(), r.token.AccessToken, r.realm, sid,
		)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		// delete the already mapped roles
		var deleteRoles []gocloak.Role
		for _, r := range existingRoles {
			deleteRoles = append(deleteRoles, *r)
		}
		err = r.client.DeleteClientScopesScopeMappingsRealmRoles(
			context.Background(), r.token.AccessToken, r.realm, sid, deleteRoles,
		)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		// create new scope mappings
		err = r.client.CreateClientScopesScopeMappingsRealmRoles(
			context.Background(), r.token.AccessToken, r.realm, sid, roles,
		)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		// ensure that the scope will be on the token by default
		err = r.client.AddDefaultScopeToClient(
			context.Background(), r.token.AccessToken, r.realm, *c.ID, sid,
		)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	return
}

//
// getClient returns a keycloak realm client.
func (r *Keycloak) getClient(clientId string) (client *gocloak.Client, found bool, err error) {
	max := 1
	clientParams := gocloak.GetClientsParams{ClientID: &clientId, Max: &max}
	clients, err := r.client.GetClients(
		context.Background(), r.token.AccessToken, r.realm, clientParams,
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if len(clients) == 0 {
		return
	}

	found = true
	client = clients[0]
	return
}

//
// userMap generates a mapping of usernames to user objects
func (r *Keycloak) userMap() (userMap map[string]gocloak.User, err error) {
	userMap = make(map[string]gocloak.User)
	users, err := r.client.GetUsers(
		context.Background(), r.token.AccessToken, r.realm, gocloak.GetUsersParams{},
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, user := range users {
		userMap[*user.Username] = *user
	}
	return
}

//
// realmRoleMap generates a mapping of realm role names to role objects
func (r *Keycloak) realmRoleMap() (roleMap map[string]gocloak.Role, err error) {
	roleMap = make(map[string]gocloak.Role)
	realmRoles, err := r.client.GetRealmRoles(
		context.Background(), r.token.AccessToken, r.realm, gocloak.GetRoleParams{},
	)
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
// scopeMap generates a mapping of client scope names to scope objects
func (r *Keycloak) scopeMap() (scopeMap map[string]gocloak.ClientScope, err error) {
	scopeMap = make(map[string]gocloak.ClientScope)
	clientScopes, err := r.client.GetClientScopes(
		context.Background(), r.token.AccessToken, r.realm,
	)
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
func (r *Keycloak) login() (err error) {
	// retry for three minutes to allow for the possibility
	// that the hub came up before keycloak did.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			r.token, err = r.client.LoginAdmin(ctx, r.admin, r.pass, r.adminRealm)
			if err != nil {
				log.Info("Login failed.", "reason", err.Error())
				time.Sleep(time.Second)
			} else {
				return
			}
		}
	}
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
