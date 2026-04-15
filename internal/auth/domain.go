package auth

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/auth/seed"
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
	registeredScopes = make(map[string]bool)
)

func init() {
	var err error
	seedDir, err = fs.Sub(seedFS, "seed")
	if err != nil {
		panic(err)
	}
}

// RegisterScope registers a resource scope for permission generation.
func RegisterScope(scope string) {
	registeredScopes[scope] = true
}

// NewDomain returns a domain.
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

// Seed seeds permissions, roles, and users.
func (d *Domain) Seed() (err error) {
	database.PK.Begin(d.DB, model.Permission{}, 1000)
	database.PK.Begin(d.DB, model.Role{}, 1000)
	database.PK.Begin(d.DB, model.User{}, 1000)
	var resources []string
	for scope := range registeredScopes {
		resources = append(resources, scope)
	}
	sort.Strings(resources)
	err = d.seedPermissions(resources)
	if err != nil {
		return
	}
	err = d.seedRoles()
	if err != nil {
		return
	}
	err = d.buildRoleMap()
	if err != nil {
		return
	}
	err = d.seedUsers()
	return
}

// seedPermissions seeds permissions based on discovered route scopes.
// Preserves existing permission IDs, deletes orphaned permissions,
// and adds new permissions with sequential IDs.
// Builds an in-memory map of scope to permission ID for role seeding.
func (d *Domain) seedPermissions(resources []string) (err error) {
	perms := d.generatePermissions(resources)
	err = d.reconcilePermissions(perms)
	if err != nil {
		return
	}
	err = d.buildPermissionMap()
	return
}

// buildPermissionMap reads all permissions and builds scope->ID map.
func (d *Domain) buildPermissionMap() (err error) {
	permDefs, err := d.readPermissions(d.DB)
	if err != nil {
		return
	}
	d.permByScope = make(map[string]uint)
	for scope, perm := range permDefs {
		d.permByScope[scope] = perm.ID
	}
	return
}

// seedRoles seeds roles and their permission associations from roles.yaml.
// Must be called after seedPermissions to ensure permission map is populated.
// Preserves existing role IDs, deletes orphaned seeded roles (ID < 1000),
// and creates new roles with static IDs from YAML.
func (d *Domain) seedRoles() (err error) {
	roles, err := d.readRoles()
	if err != nil {
		return
	}
	err = d.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			existing, err := d.readExistingRoles(tx)
			if err != nil {
				return
			}
			toDelete, toUpdate, toCreate := d.diffRoles(existing, roles)
			err = d.deleteRoles(tx, toDelete)
			if err != nil {
				return
			}
			for _, role := range toUpdate {
				err = d.updateRole(tx, existing[role.Name], role)
				if err != nil {
					return
				}
			}
			for _, role := range toCreate {
				err = d.createRole(tx, role)
				if err != nil {
					return
				}
			}
			return
		})
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

// readExistingRoles reads existing roles from database.
func (d *Domain) readExistingRoles(db *gorm.DB) (roles map[string]model.Role, err error) {
	var list []model.Role
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	roles = make(map[string]model.Role)
	for _, r := range list {
		roles[r.Name] = r
	}
	return
}

// diffRoles calculates which roles to delete, update, or create.
func (d *Domain) diffRoles(existing map[string]model.Role, wanted []seed.Role) (
	toDelete []uint, toUpdate []seed.Role, toCreate []seed.Role) {
	wantedMap := make(map[string]seed.Role)
	for _, role := range wanted {
		wantedMap[role.Name] = role
	}
	for name, role := range existing {
		if role.ID < 1000 {
			if _, found := wantedMap[name]; !found {
				toDelete = append(toDelete, role.ID)
			}
		}
	}
	for _, role := range wanted {
		if _, found := existing[role.Name]; found {
			toUpdate = append(toUpdate, role)
		} else {
			toCreate = append(toCreate, role)
		}
	}
	return
}

// deleteRoles deletes roles with the given IDs.
func (d *Domain) deleteRoles(db *gorm.DB, ids []uint) (err error) {
	if len(ids) == 0 {
		return
	}
	err = db.Delete(&model.Role{}, "id IN ?", ids).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// updateRole updates an existing role's permission associations.
func (d *Domain) updateRole(db *gorm.DB, existing model.Role, role seed.Role) (err error) {
	perms, err := d.buildRolePermissions(role)
	if err != nil {
		return
	}
	err = db.Model(&existing).Association("Permissions").Replace(perms)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// createRole creates a new role with permissions.
func (d *Domain) createRole(db *gorm.DB, role seed.Role) (err error) {
	perms, err := d.buildRolePermissions(role)
	if err != nil {
		return
	}
	m := &model.Role{
		Name: role.Name,
	}
	m.ID = role.ID
	err = db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Model(m).Association("Permissions").Replace(perms)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// buildRolePermissions builds permission list from a role definition.
func (d *Domain) buildRolePermissions(role seed.Role) (perms []model.Permission, err error) {
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
		perms = append(perms, model.Permission{
			Model: model.Model{ID: permID},
		})
	}
	return
}

// buildRoleMap reads all roles and builds name->ID map.
func (d *Domain) buildRoleMap() (err error) {
	var list []model.Role
	err = d.DB.Find(&list).Error
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
// Preserves existing user IDs, deletes orphaned seeded users (ID < 1000),
// and creates new users with static IDs from YAML.
func (d *Domain) seedUsers() (err error) {
	users, err := d.readUsers()
	if err != nil {
		return
	}
	err = d.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			existing, err := d.readExistingUsers(tx)
			if err != nil {
				return
			}
			toDelete, toUpdate, toCreate := d.diffUsers(existing, users)
			err = d.deleteUsers(tx, toDelete)
			if err != nil {
				return
			}
			for _, user := range toUpdate {
				err = d.updateUser(tx, existing[user.Userid], user)
				if err != nil {
					return
				}
			}
			for _, user := range toCreate {
				err = d.createUser(tx, user)
				if err != nil {
					return
				}
			}
			return
		})
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

// readExistingUsers reads existing users from database.
func (d *Domain) readExistingUsers(db *gorm.DB) (users map[string]model.User, err error) {
	var list []model.User
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	users = make(map[string]model.User)
	for _, u := range list {
		users[u.Userid] = u
	}
	return
}

// diffUsers calculates which users to delete, update, or create.
func (d *Domain) diffUsers(existing map[string]model.User, wanted []seed.User) (
	toDelete []uint, toUpdate []seed.User, toCreate []seed.User) {
	wantedMap := make(map[string]seed.User)
	for _, user := range wanted {
		wantedMap[user.Userid] = user
	}
	for userid, user := range existing {
		if user.ID < 1000 {
			if _, found := wantedMap[userid]; !found {
				toDelete = append(toDelete, user.ID)
			}
		}
	}
	for _, user := range wanted {
		if _, found := existing[user.Userid]; found {
			toUpdate = append(toUpdate, user)
		} else {
			toCreate = append(toCreate, user)
		}
	}
	return
}

// deleteUsers deletes users with the given IDs.
func (d *Domain) deleteUsers(db *gorm.DB, ids []uint) (err error) {
	if len(ids) == 0 {
		return
	}
	err = db.Delete(&model.User{}, "id IN ?", ids).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// updateUser updates an existing user's role associations and password.
func (d *Domain) updateUser(db *gorm.DB, existing model.User, user seed.User) (err error) {
	roles, err := d.buildUserRoles(user)
	if err != nil {
		return
	}
	err = db.Model(&existing).Association("Roles").Replace(roles)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// createUser creates a new user with roles.
func (d *Domain) createUser(db *gorm.DB, user seed.User) (err error) {
	roles, err := d.buildUserRoles(user)
	if err != nil {
		return
	}
	m := &model.User{
		Userid:   user.Userid,
		Subject:  uuid.New().String(),
		Password: secret.HashPassword(user.Password),
	}
	m.ID = user.ID
	err = db.Create(m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Model(m).Association("Roles").Replace(roles)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// buildUserRoles builds role list from a user definition.
func (d *Domain) buildUserRoles(user seed.User) (roles []model.Role, err error) {
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
		roles = append(roles, model.Role{
			Model: model.Model{ID: roleID},
		})
	}
	return
}

// generatePermissions generates all permissions for the given resources.
// Each resource gets 5 permissions (one per HTTP verb).
func (d *Domain) generatePermissions(resources []string) (perms []model.Permission) {
	verbs := []string{
		"delete",
		"get",
		"patch",
		"post",
		"put",
	}
	for _, resource := range resources {
		for _, verb := range verbs {
			scope := resource + ":" + verb
			perms = append(perms, model.Permission{
				Name:  scope,
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

// reconcilePermissions reconcilehronizes permissions in the database with the wanted set.
// Preserves existing permission IDs, deletes orphaned permissions,
// and assigns sequential IDs to new permissions.
func (d *Domain) reconcilePermissions(wanted []model.Permission) (err error) {
	err = d.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			existing, err := d.readPermissions(tx)
			if err != nil {
				return
			}
			nextID := d.maxID(existing) + 1
			toDelete, toCreate := d.diffPermissions(existing, wanted, &nextID)
			err = d.deletePermissions(tx, toDelete)
			if err != nil {
				return
			}
			err = d.createPermissions(tx, toCreate)
			return
		})
	if err != nil {
		err = liberr.Wrap(err)
	}

	return
}

// readPermissions reads all existing permissions from the database.
func (d *Domain) readPermissions(db *gorm.DB) (perms map[string]model.Permission, err error) {
	var list []model.Permission
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	perms = make(map[string]model.Permission)
	for _, p := range list {
		perms[p.Scope] = p
	}
	return
}

// maxID returns the maximum ID from a set of permissions.
func (d *Domain) maxID(perms map[string]model.Permission) (max uint) {
	for _, p := range perms {
		if p.ID > max {
			max = p.ID
		}
	}
	return
}

// diffPermissions calculates which permissions to delete and create.
// Assigns IDs to new permissions starting from nextID.
func (d *Domain) diffPermissions(existing map[string]model.Permission, wanted []model.Permission, nextID *uint) (
	toDelete []string, toCreate []model.Permission) {
	wantedSet := make(map[string]model.Permission)
	for _, perm := range wanted {
		wantedSet[perm.Scope] = perm
	}
	for scope := range existing {
		if _, found := wantedSet[scope]; !found {
			toDelete = append(toDelete, scope)
		}
	}
	for _, perm := range wanted {
		if _, found := existing[perm.Scope]; found {
			continue
		}
		perm.ID = *nextID
		*nextID++
		toCreate = append(toCreate, perm)
	}
	return
}

// deletePermissions deletes permissions with the given scopes.
func (d *Domain) deletePermissions(db *gorm.DB, scopes []string) (err error) {
	if len(scopes) == 0 {
		return
	}
	err = db.Delete(&model.Permission{}, "scope IN ?", scopes).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// createPermissions creates new permissions in the database.
func (d *Domain) createPermissions(db *gorm.DB, perms []model.Permission) (err error) {
	for _, perm := range perms {
		err = db.Create(&perm).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

type ScopeNotFound struct {
	Scope string
}

func (e *ScopeNotFound) Error() string {
	return fmt.Sprintf("Scope %s not-found.", e.Scope)
}

func (e *ScopeNotFound) Is(err error) (matched bool) {
	var inst *ScopeNotFound
	matched = errors.As(err, &inst)
	return
}
