package auth

import (
	"embed"
	"io/fs"
	"sort"

	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/auth/seed"
	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var (
	//go:embed seed
	seedFS  embed.FS
	seedDir fs.FS
)

const (
	LastId = 1000
)

var Tenant *Domain

func init() {
	Tenant = NewDomain(nil)
	var err error
	seedDir, err = fs.Sub(seedFS, "seed")
	if err != nil {
		panic(err)
	}
}

// NewDomain returns a new RBAC domain manager.
func NewDomain(db *gorm.DB) *Domain {
	return &Domain{
		DB:          db,
		roleByName:  make(map[string]uint),
		scopeByName: make(map[string]Scope),
		resources: map[string]bool{
			ADMIN: true,
		},
	}
}

// Domain the RBAC domain.
type Domain struct {
	DB          *gorm.DB
	resources   map[string]bool
	roleByName  map[string]uint
	scopeByName map[string]Scope
}

// Register registers a scope resource.
func (d *Domain) Register(resource string) {
	if resource != "" {
		d.resources[resource] = true
	}
}

// Resources returns a list of registered resources.
func (d *Domain) Resources() (resources []string) {
	for resource := range d.resources {
		resources = append(resources, resource)
	}
	sort.Strings(resources)
	return
}

func (d *Domain) Scopes() (scopes []string) {
	for s := range d.scopeByName {
		scopes = append(scopes, s)
	}
	sort.Strings(scopes)
	return
}

// Seed seeds roles, clients, and users.
func (d *Domain) Seed() (err error) {
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
func (d *Domain) buildScopes() {
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

// seedRoles seeds roles from roles.yaml.
// Must be called after buildScopes to ensure scopes map is populated.
// Preserves existing role IDs, deletes orphaned seeded roles (ID < MaxId),
// and creates new roles with static IDs from YAML.
func (d *Domain) seedRoles(db *gorm.DB) (err error) {
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

// readRoles reads role definitions from roles.yaml.
func (d *Domain) readRoles() (roles []seed.Role, err error) {
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

// fetchRoles fetches existing roles from database.
func (d *Domain) fetchRoles(db *gorm.DB) (roles map[string]Role, err error) {
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

// rolePatch computes the role reconciliation patch.
func (d *Domain) rolePatch(existing map[string]Role, wanted []seed.Role) (patch *RolePatch) {
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

// buildRoleScopes builds scope strings from a role definition.
func (d *Domain) buildRoleScopes(role seed.Role) (scopes []string) {
	scopeSet := make(map[string]bool)
	for _, r := range role.Resources {
		for _, m := range r.Verbs {
			scope := Scope{Resource: r.Name, Method: m}
			for _, s := range scope.ExpandWith(d.Resources()) {
				scopeStr := s.String()
				if _, found := d.scopeByName[scopeStr]; !found {
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

// buildRoleMap reads all roles and builds name->ID map.
func (d *Domain) buildRoleMap(db *gorm.DB) (err error) {
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

// seedUsers seeds users and their role associations from users.yaml.
// Must be called after seedRoles to ensure role map is populated.
// Preserves existing user IDs, deletes orphaned seeded users (ID < MaxId),
// and creates new users with static IDs from YAML.
func (d *Domain) seedUsers(db *gorm.DB) (err error) {
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

// readUsers reads user definitions from users.yaml.
func (d *Domain) readUsers() (users []seed.User, err error) {
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

// fetchUsers fetches existing users from database.
func (d *Domain) fetchUsers(db *gorm.DB) (users map[string]User, err error) {
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

// userPatch computes the user reconciliation patch.
func (d *Domain) userPatch(existing map[string]User, wanted []seed.User) (patch *UserPatch) {
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

// buildUserRoles builds role list from a user definition.
func (d *Domain) buildUserRoles(user seed.User) (roles []Role, err error) {
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

// UserPatch represents changes to reconcile users.
type UserPatch struct {
	db       *gorm.DB
	toDelete []uint
	toUpdate []userWithRoles
	toCreate []userWithRoles
}

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

// seedClients seeds OIDC clients from CRDs.
// Preserves existing client IDs, deletes orphaned seeded clients (ID < LastId),
// and creates new clients with IDs from CRD spec.
func (d *Domain) seedClients(db *gorm.DB) (err error) {
	existing, err := d.fetchClients(db)
	if err != nil {
		return
	}
	patch := d.clientPatch(existing, federated.Clients)
	err = patch.Apply(db)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// fetchClients fetches existing clients from database.
func (d *Domain) fetchClients(db *gorm.DB) (clients map[string]IdpClient, err error) {
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

// clientPatch computes the client reconciliation patch from CRD clients.
func (d *Domain) clientPatch(existing map[string]IdpClient, wanted []as.IdpClient) (patch *IdpClientPatch) {
	patch = &IdpClientPatch{
		db: d.DB,
	}
	wantedMap := make(map[string]as.IdpClient)
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
	for _, settingsClient := range wanted {
		if existingClient, found := existing[settingsClient.ClientId]; found {
			patch.toUpdate = append(patch.toUpdate, clientWithSettings{
				client:   existingClient,
				settings: settingsClient,
			})
		} else {
			newClient := IdpClient{}
			newClient.With(&settingsClient)
			newClient.Subject = uuid.New().String()
			patch.toCreate = append(patch.toCreate, clientWithSettings{
				client:   newClient,
				settings: settingsClient,
			})
		}
	}
	return
}

// IdpClientPatch represents changes to reconcile clients from YAML.
type IdpClientPatch struct {
	db       *gorm.DB
	toDelete []uint
	toUpdate []clientWithSettings
	toCreate []clientWithSettings
}

// clientWithSettings pairs database client with settings data.
type clientWithSettings struct {
	client   IdpClient
	settings as.IdpClient
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
	for _, item := range p.toUpdate {
		m := &IdpClient{}
		m.With(&item.settings)
		m.ID = item.client.ID
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
