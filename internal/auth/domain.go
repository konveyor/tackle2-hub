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

var (
	registeredResources = map[string]bool{
		ADMIN: true,
	}
)

const (
	LastId = 1000
)

func init() {
	var err error
	seedDir, err = fs.Sub(seedFS, "seed")
	if err != nil {
		panic(err)
	}
}

// RegisterResource registers an api resource for permission (scope) generation.
func RegisterResource(resource string) {
	if resource != "" {
		registeredResources[resource] = true
	}
}

// NewDomain returns a new RBAC domain manager.
func NewDomain(db *gorm.DB) *Domain {
	return &Domain{
		DB:          db,
		permByScope: make(map[string]uint),
		roleByName:  make(map[string]uint),
	}
}

// Domain the RBAC domain.
type Domain struct {
	DB          *gorm.DB
	permByScope map[string]uint
	roleByName  map[string]uint
}

// Seed seeds permissions, roles, clients, and users.
func (d *Domain) Seed() (err error) {
	database.PK.Begin(d.DB, Permission{}, LastId)
	database.PK.Begin(d.DB, Role{}, LastId)
	database.PK.Begin(d.DB, IdpClient{}, LastId)
	database.PK.Begin(d.DB, User{}, LastId)
	var resources []string
	for r := range registeredResources {
		resources = append(resources, r)
	}
	sort.Strings(resources)
	err = d.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			err = d.seedPermissions(tx, resources)
			if err != nil {
				return
			}
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

// seedPermissions seeds permissions based on discovered route scopes.
// Preserves existing permission IDs, deletes orphaned permissions,
// and adds new permissions with sequential IDs.
// Builds an in-memory map of scope to permission ID for role seeding.
func (d *Domain) seedPermissions(db *gorm.DB, resources []string) (err error) {
	perms := d.generatePermissions(resources)
	err = d.reconcilePermissions(db, perms)
	if err != nil {
		return
	}
	err = d.buildPermissionMap(db)
	return
}

// buildPermissionMap reads all permissions and builds scope->ID map.
func (d *Domain) buildPermissionMap(db *gorm.DB) (err error) {
	permDefs, err := d.fetchPermissions(db)
	if err != nil {
		return
	}
	d.permByScope = make(map[string]uint)
	for scope, perm := range permDefs {
		d.permByScope[scope] = perm.ID
	}
	return
}

// generatePermissions generates all permissions for the given resources.
// Each resource gets 5 permissions (one per HTTP verb).
func (d *Domain) generatePermissions(resources []string) (perms []Permission) {
	verbs := []string{
		"decrypt",
		"delete",
		"get",
		"patch",
		"post",
		"put",
	}
	for _, resource := range resources {
		for _, verb := range verbs {
			name := verb + "-" + resource
			scope := resource + ":" + verb
			perms = append(perms, Permission{
				Name:  name,
				Scope: scope,
			})
		}
	}
	sort.Slice(
		perms,
		func(i, j int) bool {
			return perms[i].Scope < perms[j].Scope
		})
	return
}

// reconcilePermissions reconcile permissions in the database with the wanted set.
// Preserves existing permission IDs, deletes orphaned permissions,
// and assigns sequential IDs to new permissions.
func (d *Domain) reconcilePermissions(db *gorm.DB, wanted []Permission) (err error) {
	existing, err := d.fetchPermissions(db)
	if err != nil {
		return
	}
	nextID := d.maxID(existing) + 1
	patch := d.permissionPatch(existing, wanted, nextID)
	err = patch.Apply(db)
	return
}

// fetchPermissions fetches all existing permissions from the database.
func (d *Domain) fetchPermissions(db *gorm.DB) (perms map[string]Permission, err error) {
	var list []Permission
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	perms = make(map[string]Permission)
	for _, p := range list {
		perms[p.Scope] = p
	}
	return
}

// maxID returns the maximum ID from a set of permissions.
func (d *Domain) maxID(perms map[string]Permission) (max uint) {
	for _, p := range perms {
		if p.ID > max {
			max = p.ID
		}
	}
	return
}

// permissionPatch computes the permission reconciliation patch.
func (d *Domain) permissionPatch(existing map[string]Permission, wanted []Permission, nextID uint) (patch *PermissionPatch) {
	patch = &PermissionPatch{
		db: d.DB,
	}
	wantedSet := make(map[string]Permission)
	for _, perm := range wanted {
		wantedSet[perm.Scope] = perm
	}
	for scope := range existing {
		if _, found := wantedSet[scope]; !found {
			patch.toDelete = append(patch.toDelete, scope)
		}
	}
	for _, perm := range wanted {
		if _, found := existing[perm.Scope]; found {
			continue
		}
		perm.ID = nextID
		nextID++
		patch.toCreate = append(patch.toCreate, perm)
	}
	return
}

// seedRoles seeds roles and their permission associations from roles.yaml.
// Must be called after seedPermissions to ensure permission map is populated.
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
		perms, err := d.buildRolePermissions(role)
		if err != nil {
			continue
		}
		if existingRole, found := existing[role.Name]; found {
			patch.toUpdate = append(patch.toUpdate, roleWithPermissions{
				role:        existingRole,
				permissions: perms,
			})
		} else {
			newRole := Role{
				Name: role.Name,
			}
			newRole.ID = role.ID
			patch.toCreate = append(patch.toCreate, roleWithPermissions{
				role:        newRole,
				permissions: perms,
			})
		}
	}
	return
}

// buildRolePermissions builds permission list from a role definition.
func (d *Domain) buildRolePermissions(role seed.Role) (perms []Permission, err error) {
	permMap := make(map[uint]bool)
	for _, resource := range role.Resources {
		for _, verb := range resource.Verbs {
			scope := resource.Name + ":" + verb
			permID, found := d.permByScope[scope]
			if !found {
				Log.Info(
					"Role has unknown scope.",
					"name",
					role.Name,
					"scope",
					scope)
				continue
			}
			permMap[permID] = true
		}
	}
	for permID := range permMap {
		perms = append(perms, Permission{
			Model: Model{ID: permID},
		})
	}
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
			newUser := User{
				Login:    user.Login,
				Subject:  uuid.New().String(),
				Password: user.Password,
			}
			newUser.ID = user.ID
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
	toUpdate []roleWithPermissions
	toCreate []roleWithPermissions
}

type roleWithPermissions struct {
	role        Role
	permissions []Permission
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
		item := &p.toUpdate[i]
		perms := make([]model.Permission, len(item.permissions))
		for j := range item.permissions {
			perms[j] = item.permissions[j]
		}
		err = db.Model(&item.role).Association("Permissions").Replace(perms)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for i := range p.toCreate {
		item := &p.toCreate[i]
		err = db.Create(&item.role).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		perms := make([]model.Permission, len(item.permissions))
		for j := range item.permissions {
			perms[j] = item.permissions[j]
		}
		err = db.Model(&item.role).Association("Permissions").Replace(perms)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// PermissionPatch represents changes to reconcile permissions.
type PermissionPatch struct {
	db       *gorm.DB
	toDelete []string
	toCreate []Permission
}

// Apply applies the permission patch to the database.
func (p *PermissionPatch) Apply(db *gorm.DB) (err error) {
	if len(p.toDelete) > 0 {
		err = db.Delete(&Permission{}, "scope IN ?", p.toDelete).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	for _, perm := range p.toCreate {
		err = db.Create(&perm).Error
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
