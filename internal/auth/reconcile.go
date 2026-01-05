package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	liberr "github.com/jortel/go-utils/error"
)

// NewReconciler builds a new Keycloak realm reconciler.
func NewReconciler(host, realm, id, secret, admin, pass, adminRealm string) (r Reconciler) {
	client := gocloak.NewClient(host, gocloak.SetAuthRealms("auth/realms"), gocloak.SetAuthAdminRealms("auth/admin/realms"))
	client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	r = Reconciler{
		client:     client,
		realm:      realm,
		id:         id,
		secret:     secret,
		admin:      admin,
		pass:       pass,
		adminRealm: adminRealm,
	}
	return
}

// Keycloak realm reconciler
type Reconciler struct {
	client     *gocloak.GoCloak
	realm      string
	id         string
	secret     string
	admin      string
	pass       string
	adminRealm string
	token      *gocloak.JWT
}

// Realm is a container for the users,
// scopes, and roles that exist in the
// hub's keycloak realm.
type Realm struct {
	Users  map[string]gocloak.User
	Scopes map[string]gocloak.ClientScope
	Roles  map[string]gocloak.Role
}

// Reconcile ensures that the Hub realm
// exists and the expected clients, roles, scopes,
// and users are present in it.
func (r *Reconciler) Reconcile() (err error) {
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

	Log.Info("Realm synced.")

	return
}

// loadRealm loads all the scopes, roles, and users from
// the hub keycloak realm and populates a Realm struct.
func (r *Reconciler) loadRealm() (realm *Realm, err error) {
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

// ensureRealm ensures that the hub realm exists.
func (r *Reconciler) ensureRealm() (err error) {
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
			Log.Info("Creating realm.", "realm", r.realm)
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

// ensureClient ensures that the hub client exists.
func (r *Reconciler) ensureClient() (err error) {
	var found bool
	var existingClient *gocloak.Client
	existingClient, found, err = r.getClient(r.id)
	if err != nil {
		return
	}
	if found {
		err = r.ensureAudienceMapper(existingClient)
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
	Log.Info("Creating client.", "client", r.id)
	clientID, err := r.client.CreateClient(context.Background(), r.token.AccessToken, r.realm, newClient)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	createdClient, err := r.client.GetClient(context.Background(), r.token.AccessToken, r.realm, clientID)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = r.ensureAudienceMapper(createdClient)
	return
}

func (r *Reconciler) ensureAudienceMapper(client *gocloak.Client) (err error) {
	if Settings.Auth.Keycloak.Audience == "" {
		Log.Info("Skipping audience mapper creation - no audience configured")
		return
	}

	mapperName := "audience-mapper"
	if client.ProtocolMappers != nil {
		for _, mapper := range *client.ProtocolMappers {
			if mapper.Name != nil && *mapper.Name == mapperName {
				Log.Info("Audience mapper already exists.", "client", *client.ClientID)
				return
			}
		}
	}

	protocol := "openid-connect"
	protocolMapper := "oidc-audience-mapper"
	mapper := gocloak.ProtocolMapperRepresentation{
		Name:           &mapperName,
		Protocol:       &protocol,
		ProtocolMapper: &protocolMapper,
		Config: &map[string]string{
			"included.custom.audience": Settings.Auth.Keycloak.Audience,
			"access.token.claim":       "true",
			"id.token.claim":           "true",
		},
	}

	_, err = r.client.CreateClientProtocolMapper(
		context.Background(),
		r.token.AccessToken,
		r.realm,
		*client.ID,
		mapper,
	)
	if err != nil {
		Log.Error(err, "Failed to create audience mapper", "client", *client.ClientID, "audience", Settings.Auth.Keycloak.Audience)
		err = liberr.Wrap(err)
		return
	}

	Log.Info("Created audience mapper.", "client", *client.ClientID, "audience", Settings.Auth.Keycloak.Audience)
	return
}

// ensureUsers ensures that the hub users exist and have the necessary roles.
func (r *Reconciler) ensureUsers(realm *Realm) (err error) {
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
				Username:  &user.Name,
				Enabled:   &enabled,
				FirstName: &user.FirstName,
				LastName:  &user.LastName,
				Email:     &user.Email,
			}
			if Settings.Keycloak.RequirePasswordUpdate {
				u.RequiredActions = &[]string{"UPDATE_PASSWORD"}
			}
			Log.Info("Creating user.", "user", user.Name)
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
			Log.Info("Removing any existing roles from user.", "user", user.Name)
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
		Log.Info("Applying roles to user.", "user", user.Name, "roles", user.Roles)
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

// ensureRoles ensures that hub roles and scopes are present in keycloak by
// creating them if they are missing and assigning scope mappings.
func (r *Reconciler) ensureRoles(realm *Realm) (err error) {
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
			Log.Info("Created realm role.", "role", role.Name, "realm", r.realm)
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
					Log.Info("Created client scope.", "scope", scopeName, "realm", r.realm)
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
		Log.Info("Syncing scope mappings.", "scope", sid, "roles", roles)
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

// getClient returns a keycloak realm client.
func (r *Reconciler) getClient(clientId string) (client *gocloak.Client, found bool, err error) {
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

// userMap generates a mapping of usernames to user objects
func (r *Reconciler) userMap() (userMap map[string]gocloak.User, err error) {
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

// realmRoleMap generates a mapping of realm role names to role objects
func (r *Reconciler) realmRoleMap() (roleMap map[string]gocloak.Role, err error) {
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

// scopeMap generates a mapping of client scope names to scope objects
func (r *Reconciler) scopeMap() (scopeMap map[string]gocloak.ClientScope, err error) {
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

// login logs into the keycloak admin-cli client as the administrator.
func (r *Reconciler) login() (err error) {
	for {
		r.token, err = r.client.LoginAdmin(context.Background(), r.admin, r.pass, r.adminRealm)
		if err != nil {
			Log.Info("Login failed.", "reason", err.Error())
			time.Sleep(time.Second * 3)
		} else {
			return
		}
	}
}
