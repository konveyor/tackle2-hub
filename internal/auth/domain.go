package auth

import (
	"context"
	"crypto/tls"
	"embed"
	"io/fs"
	"net/url"
	"sort"
	"strings"

	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/auth/seed"
	"github.com/konveyor/tackle2-hub/internal/database"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	//go:embed seed
	seedFS  embed.FS
	seedDir fs.FS
)

const (
	LastId = 1000
)

// Domain is the default tenant.
var Domain *Tenant

func init() {
	Domain = NewTenant(nil, nil)
	var err error
	seedDir, err = fs.Sub(seedFS, "seed")
	if err != nil {
		panic(err)
	}
}

// NewTenant returns a new RBAC domain manager.
func NewTenant(db *gorm.DB, client k8sClient.Client) *Tenant {
	return &Tenant{
		DB:          db,
		client:      client,
		roleByName:  make(map[string]uint),
		scopeByName: make(map[string]Scope),
		resources: map[string]bool{
			ADMIN: true,
		},
	}
}

// Tenant the RBAC domain.
type Tenant struct {
	DB          *gorm.DB
	client      k8sClient.Client
	resources   map[string]bool
	roleByName  map[string]uint
	scopeByName map[string]Scope
	Idp         IdentityProvider
	Ldap        LdapProvider
}

// Register registers a scope resource.
func (d *Tenant) Register(resource string) {
	if resource != "" {
		d.resources[resource] = true
	}
}

// Resources returns a list of registered resources.
func (d *Tenant) Resources() (resources []string) {
	for resource := range d.resources {
		resources = append(resources, resource)
	}
	sort.Strings(resources)
	return
}

// Scopes returns the list of known scopes.
func (d *Tenant) Scopes() (scopes []string) {
	for s := range d.scopeByName {
		scopes = append(scopes, s)
	}
	sort.Strings(scopes)
	return
}

// HasScope returns true when the domain has the scope.
func (d *Tenant) HasScope(scope string) (found bool) {
	_, found = d.scopeByName[scope]
	return
}

// Load the domain.
func (d *Tenant) Load() (err error) {
	err = d.getIdp()
	if err != nil {
		return
	}
	err = d.getLdap()
	if err != nil {
		return
	}
	return
}

// Seed seeds roles, clients, and users.
func (d *Tenant) Seed() (err error) {
	database.PK.Begin(d.DB, Role{}, LastId)
	database.PK.Begin(d.DB, IdpClient{}, LastId)
	database.PK.Begin(d.DB, User{}, LastId)
	var resources []string
	for r := range d.resources {
		resources = append(resources, r)
	}
	sort.Strings(resources)
	err = d.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			d.buildScopes()
			err = d.seedRoles(tx)
			if err != nil {
				return
			}
			err = d.pruneScopes(tx)
			if err != nil {
				return
			}
			err = d.buildRoleMap(tx)
			if err != nil {
				return
			}
			err = d.seedClients(tx)
			if err != nil {
				return
			}
			err = d.seedUsers(tx)
			return
		})
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// buildScopes builds the map of scopes.
func (d *Tenant) buildScopes() {
	for resource := range d.resources {
		for _, verb := range verbs {
			scope := Scope{
				Resource: resource,
				Method:   verb,
			}
			d.scopeByName[scope.String()] = scope
		}
	}
	return
}

// buildRoleMap reads all roles and builds name->ID map.
func (d *Tenant) buildRoleMap(db *gorm.DB) (err error) {
	var list []Role
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	d.roleByName = make(map[string]uint)
	for _, role := range list {
		d.roleByName[role.Name] = role.ID
	}
	return
}

// buildRoleScopes builds scope strings from a role definition.
func (d *Tenant) buildRoleScopes(role seed.Role) (scopes []string) {
	scopeSet := make(map[string]bool)
	for _, r := range role.Resources {
		for _, m := range r.Verbs {
			scope := Scope{Resource: r.Name, Method: m}
			for _, s := range scope.ExpandWith(d.Resources()) {
				scopeStr := s.String()
				if !d.HasScope(scopeStr) {
					Log.Info(
						"Role has unknown scope.",
						"name",
						role.Name,
						"scope",
						scopeStr)
					continue
				}
				scopeSet[scopeStr] = true
			}
		}
	}
	for scopeStr := range scopeSet {
		scopes = append(scopes, scopeStr)
	}
	sort.Strings(scopes)
	return
}

// buildUserRoles builds role list from a user definition.
func (d *Tenant) buildUserRoles(user seed.User) (roles []Role, err error) {
	roleMap := make(map[uint]bool)
	for _, roleName := range user.Roles {
		roleID, found := d.roleByName[roleName]
		if !found {
			Log.Info("Role not-found: " + roleName)
			continue
		}
		roleMap[roleID] = true
	}
	for roleID := range roleMap {
		roles = append(roles, Role{
			Model: Model{ID: roleID},
		})
	}
	return
}

// clientPatch computes the client reconciliation patch from CRD clients.
func (d *Tenant) clientPatch(existing map[string]IdpClient, wanted []IdpClient) (patch *IdpClientPatch) {
	patch = &IdpClientPatch{
		db: d.DB,
	}
	wantedMap := make(map[string]IdpClient)
	for _, client := range wanted {
		wantedMap[client.ClientId] = client
	}
	for clientId, client := range existing {
		if client.ID < LastId {
			if _, found := wantedMap[clientId]; !found {
				patch.toDelete = append(patch.toDelete, client.ID)
			}
		}
	}
	for _, resource := range wanted {
		if existingClient, found := existing[resource.ClientId]; found {
			patch.toUpdate = append(patch.toUpdate, clientWithResource{
				client:   existingClient,
				resource: resource,
			})
		} else {
			newClient := resource
			newClient.Subject = uuid.New().String()
			patch.toCreate = append(patch.toCreate, clientWithResource{
				client:   newClient,
				resource: resource,
			})
		}
	}
	return
}

// fetchClients fetches existing clients from database.
func (d *Tenant) fetchClients(db *gorm.DB) (clients map[string]IdpClient, err error) {
	var list []IdpClient
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	clients = make(map[string]IdpClient)
	for _, c := range list {
		clients[c.ClientId] = c
	}
	return
}

// fetchRoles fetches existing roles from database.
func (d *Tenant) fetchRoles(db *gorm.DB) (roles map[string]Role, err error) {
	var list []Role
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	roles = make(map[string]Role)
	for _, r := range list {
		roles[r.Name] = r
	}
	return
}

// fetchUsers fetches existing users from database.
func (d *Tenant) fetchUsers(db *gorm.DB) (users map[string]User, err error) {
	var list []User
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	users = make(map[string]User)
	for _, u := range list {
		users[u.Login] = u
	}
	return
}

// pruneScopes removes unknown scopes from user-created roles.
// Runs during seeding after scope generation to ensure custom roles
// don't reference obsolete or unregistered scopes.
func (d *Tenant) pruneScopes(tx *gorm.DB) (err error) {
	roles := []Role{}
	err = tx.Find(&roles, "id > ?", LastId).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, role := range roles {
		kept := []string{}
		pruned := false
		for _, scope := range role.Scopes {
			if !d.HasScope(scope) {
				pruned = true
				Log.Info("Unknown scope pruned",
					"role",
					role.Name,
					"scope",
					scope)
			} else {
				kept = append(kept, scope)
			}
		}
		if pruned {
			role.Scopes = kept
			err = tx.Save(&role).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}

// readRoles reads role definitions from roles.yaml.
func (d *Tenant) readRoles() (roles []seed.Role, err error) {
	b, err := fs.ReadFile(seedDir, "roles.yaml")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = yaml.Unmarshal(b, &roles)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// readUsers reads user definitions from users.yaml.
func (d *Tenant) readUsers() (users []seed.User, err error) {
	b, err := fs.ReadFile(seedDir, "users.yaml")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = yaml.Unmarshal(b, &users)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// rolePatch computes the role reconciliation patch.
func (d *Tenant) rolePatch(existing map[string]Role, wanted []seed.Role) (patch *RolePatch) {
	patch = &RolePatch{
		db: d.DB,
	}
	wantedMap := make(map[string]seed.Role)
	for _, role := range wanted {
		wantedMap[role.Name] = role
	}
	for name, role := range existing {
		if role.ID < LastId {
			if _, found := wantedMap[name]; !found {
				patch.toDelete = append(patch.toDelete, role.ID)
			}
		}
	}
	for _, role := range wanted {
		scopes := d.buildRoleScopes(role)
		if existingRole, found := existing[role.Name]; found {
			existingRole.Scopes = scopes
			patch.toUpdate = append(patch.toUpdate, existingRole)
		} else {
			newRole := Role{
				Name:   role.Name,
				Scopes: scopes,
			}
			newRole.ID = role.ID
			patch.toCreate = append(patch.toCreate, newRole)
		}
	}
	return
}

// seedClients seeds OIDC clients from CRDs.
// Preserves existing client IDs, deletes orphaned seeded clients (ID < LastId),
// and creates new clients with IDs from CRD spec.
func (d *Tenant) seedClients(db *gorm.DB) (err error) {
	existing, err := d.fetchClients(db)
	if err != nil {
		return
	}
	resources, err := d.getClientResources()
	patch := d.clientPatch(existing, resources)
	err = patch.Apply(db)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// seedRoles seeds roles from roles.yaml.
// Must be called after buildScopes to ensure scopes map is populated.
// Preserves existing role IDs, deletes orphaned seeded roles (ID < MaxId),
// and creates new roles with static IDs from YAML.
func (d *Tenant) seedRoles(db *gorm.DB) (err error) {
	roles, err := d.readRoles()
	if err != nil {
		return
	}
	existing, err := d.fetchRoles(db)
	if err != nil {
		return
	}
	patch := d.rolePatch(existing, roles)
	err = patch.Apply(db)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// seedUsers seeds users and their role associations from users.yaml.
// Must be called after seedRoles to ensure role map is populated.
// Preserves existing user IDs, deletes orphaned seeded users (ID < MaxId),
// and creates new users with static IDs from YAML.
func (d *Tenant) seedUsers(db *gorm.DB) (err error) {
	users, err := d.readUsers()
	if err != nil {
		return
	}
	existing, err := d.fetchUsers(db)
	if err != nil {
		return
	}
	patch := d.userPatch(existing, users)
	err = patch.Apply(db)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// userPatch computes the user reconciliation patch.
func (d *Tenant) userPatch(existing map[string]User, wanted []seed.User) (patch *UserPatch) {
	patch = &UserPatch{
		db: d.DB,
	}
	wantedMap := make(map[string]seed.User)
	for _, user := range wanted {
		wantedMap[user.Login] = user
	}
	for login, user := range existing {
		if user.ID < LastId {
			if _, found := wantedMap[login]; !found {
				patch.toDelete = append(patch.toDelete, user.ID)
			}
		}
	}
	for _, user := range wanted {
		roles, err := d.buildUserRoles(user)
		if err != nil {
			continue
		}
		if existingUser, found := existing[user.Login]; found {
			patch.toUpdate = append(patch.toUpdate, userWithRoles{
				user:  existingUser,
				roles: roles,
			})
		} else {
			newUser := User{}
			newUser.ID = user.ID
			newUser.Login = user.Login
			newUser.Subject = uuid.New().String()
			newUser.Password = user.Password
			_, _ = secret.Encode(&newUser)
			patch.toCreate = append(patch.toCreate, userWithRoles{
				user:  newUser,
				roles: roles,
			})
		}
	}
	return
}

// getClientResources returns client resources.
func (d *Tenant) getClientResources() (found []IdpClient, err error) {
	list := crd.IdpClientList{}
	opt := k8sClient.InNamespace(Settings.Namespace)
	err = d.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}
	for _, m := range list.Items {
		client := IdpClient{
			ClientId:        m.Spec.ClientId,
			ApplicationType: m.Spec.ApplicationType,
			Grants:          m.Spec.Grants,
			RedirectURIs:    m.Spec.RedirectURIs,
			Scopes:          m.Spec.Scopes,
		}
		client.ID = m.Spec.ID
		ref := m.Spec.ClientSecret
		if ref != nil {
			client.Secret, err = d.getSecret(ref, "clientSecret")
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}

		found = append(found, client)
	}
	return
}

// getIdp get a federated identity provider.
func (d *Tenant) getIdp() (err error) {
	list := crd.IdentityProviderList{}
	opt := k8sClient.InNamespace(Settings.Namespace)
	err = d.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}
	for _, m := range list.Items {
		m2 := IdentityProvider{}
		err = m2.with(&m)
		if err != nil {
			return
		}
		ref := m.Spec.ClientSecret
		if ref != nil {
			m2.ClientSecret, err = d.getSecret(ref, "clientSecret")
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
		d.Idp = m2
		break
	}
	return
}

// getLdap get a federated LDAP provider.
func (d *Tenant) getLdap() (err error) {
	list := crd.LdapProviderList{}
	opt := k8sClient.InNamespace(Settings.Namespace)
	err = d.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}
	for _, m := range list.Items {
		m2 := LdapProvider{}
		err = m2.with(&m)
		if err != nil {
			return
		}
		ref := m.Spec.Password
		if ref != nil {
			m2.Password, err = d.getSecret(ref, "password")
			if err != nil {
				return
			}
		}
		d.Ldap = m2
		break
	}
	return
}

// secret returns the secret by key.
func (d *Tenant) getSecret(ref *core.ObjectReference, key string) (s string, err error) {
	secret := &core.Secret{}
	err = d.client.Get(
		context.Background(),
		k8sClient.ObjectKey{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	b := secret.Data[key]
	s = string(b)
	return
}

//
// Patches
//

// UserPatch represents changes to reconcile users.
type UserPatch struct {
	db       *gorm.DB
	toDelete []uint
	toUpdate []userWithRoles
	toCreate []userWithRoles
}

// userWithRoles pairs database user with roles data.
type userWithRoles struct {
	user  User
	roles []Role
}

// Apply applies the user patch to the database.
func (p *UserPatch) Apply(db *gorm.DB) (err error) {
	if len(p.toDelete) > 0 {
		err = db.Delete(&User{}, "id IN ?", p.toDelete).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for i := range p.toUpdate {
		item := &p.toUpdate[i]
		roles := make([]model.Role, len(item.roles))
		for j := range item.roles {
			roles[j] = model.Role(item.roles[j])
		}
		err = db.Model(&item.user).Association("Roles").Replace(roles)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for i := range p.toCreate {
		item := &p.toCreate[i]
		err = db.Create(&item.user).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		roles := make([]model.Role, len(item.roles))
		for j := range item.roles {
			roles[j] = model.Role(item.roles[j])
		}
		err = db.Model(&item.user).Association("Roles").Replace(roles)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// RolePatch represents changes to reconcile roles.
type RolePatch struct {
	db       *gorm.DB
	toDelete []uint
	toUpdate []Role
	toCreate []Role
}

// Apply applies the role patch to the database.
func (p *RolePatch) Apply(db *gorm.DB) (err error) {
	if len(p.toDelete) > 0 {
		err = db.Delete(&Role{}, "id IN ?", p.toDelete).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for i := range p.toUpdate {
		role := &p.toUpdate[i]
		err = db.Save(role).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for i := range p.toCreate {
		role := &p.toCreate[i]
		err = db.Create(role).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// IdpClientPatch represents changes to reconcile clients from YAML.
type IdpClientPatch struct {
	db       *gorm.DB
	toDelete []uint
	toUpdate []clientWithResource
	toCreate []clientWithResource
}

// clientWithSetting pairs database client with resource.
type clientWithResource struct {
	client   IdpClient
	resource IdpClient
}

// Apply applies the client patch to the database.
func (p *IdpClientPatch) Apply(db *gorm.DB) (err error) {
	if len(p.toDelete) > 0 {
		err = db.Delete(&IdpClient{}, "id IN ?", p.toDelete).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for _, m := range p.toUpdate {
		err = db.Model(m).Updates(m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for _, item := range p.toCreate {
		err = db.Create(&item.client).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

//
// Federated
//

// IdentityProvider defines a federated IdP.
type IdentityProvider struct {
	Enabled      bool
	Primary      bool
	Name         string
	Issuer       string
	ClientId     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	TLS          *tls.Config
	//
	injected bool
}

// Inject template values:
// - ${issuer}
// - ${issuer.proto}
// - ${issuer.host}
// - ${issuer.port}
// - ${issuer.path}
// Note: not thread-safe.
func (r *IdentityProvider) Inject(issuer string) {
	if r.injected {
		return
	}
	issuerURL, _ := url.Parse(issuer)
	for _, u := range []*string{
		&r.Issuer,
		&r.RedirectURI,
	} {
		*u = strings.Replace(*u, "${issuer}", issuer, -1)
		*u = strings.Replace(*u, "${issuer.proto}", issuerURL.Scheme, -1)
		*u = strings.Replace(*u, "${issuer.host}", issuerURL.Hostname(), -1)
		*u = strings.Replace(*u, "${issuer.port}", issuerURL.Port(), -1)
		*u = strings.Replace(*u, "${issuer.path}", issuerURL.Path, -1)
	}
	r.injected = true
}

// String returns a string representation.
func (r IdentityProvider) String() (s string) {
	clone := r
	clone.TLS = nil
	b, _ := yaml.Marshal(clone)
	s = string(b)
	return
}

// with populates self with the crd.
func (r *IdentityProvider) with(idp *crd.IdentityProvider) (err error) {
	r.Enabled = true
	r.Primary = idp.Spec.Primary
	r.Name = idp.Name
	r.Issuer = idp.Spec.Issuer
	r.ClientId = idp.Spec.ClientId
	r.RedirectURI = idp.Spec.RedirectURI
	r.Scopes = idp.Spec.Scopes
	r.TLS, err = idp.Spec.TLS.AsConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// LdapProvider defines a federated LDAP directory.
type LdapProvider struct {
	Enabled      bool
	Name         string
	Kind         string
	URL          string
	BaseDN       string
	BindDN       string
	Password     string
	UserFilter   string
	GroupFilter  string
	HasMemberOf  bool
	RoleMappings []MappingRule
	TLS          *tls.Config
}

// with populates self using the crd.
func (r *LdapProvider) with(ds *crd.LdapProvider) (err error) {
	r.Enabled = true
	r.Name = ds.Name
	r.Kind = ds.Spec.Kind
	r.URL = ds.Spec.URL
	r.BaseDN = ds.Spec.BaseDN
	r.BindDN = ds.Spec.BindDN
	r.UserFilter = ds.Spec.UserFilter
	r.GroupFilter = ds.Spec.GroupFilter
	r.HasMemberOf = ds.Spec.HasMemberOf
	for _, m := range ds.Spec.RoleMappings {
		r.RoleMappings = append(r.RoleMappings, MappingRule{
			Any:   m.Any,
			And:   m.And,
			Roles: m.Roles,
		})
	}
	r.TLS, err = ds.Spec.TLS.AsConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
