package auth

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var (
	//go:embed seed
	seedFS  embed.FS
	seedDir fs.FS
)

var (
	scopeRegistry = make(map[string]bool)
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
	scopeRegistry[scope] = true
}

func NewDomain(db *gorm.DB) *Domain {
	return &Domain{
		DB:          db,
		permByScope: make(map[string]uint),
	}
}

// Role use to read roles.yaml.
type Role struct {
	ID        uint   `yaml:"id"`
	Name      string `yaml:"role"`
	Resources []struct {
		Name  string   `yaml:"name"`
		Verbs []string `yaml:"verbs"`
	} `yaml:"resources"`
}

// Domain the RBAC domain.
type Domain struct {
	DB          *gorm.DB
	permByScope map[string]uint
}

// Seed seeds both permissions and roles.
// Discovers route scopes, seeds permissions, builds permission map, then seeds roles.
func (d *Domain) Seed() (err error) {
	database.PK.Begin(d.DB, model.Permission{}, 1000)
	database.PK.Begin(d.DB, model.Role{}, 1000)
	var resources []string
	for scope := range scopeRegistry {
		resources = append(resources, scope)
	}
	sort.Strings(resources)
	err = d.seedPermissions(resources)
	if err != nil {
		return
	}
	err = d.seedRoles()
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
func (d *Domain) readRoles() (roles []Role, err error) {
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
func (d *Domain) diffRoles(existing map[string]model.Role, wanted []Role) (
	toDelete []uint, toUpdate []Role, toCreate []Role) {
	wantedMap := make(map[string]Role)
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
func (d *Domain) updateRole(db *gorm.DB, existing model.Role, role Role) (err error) {
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
func (d *Domain) createRole(db *gorm.DB, role Role) (err error) {
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
func (d *Domain) buildRolePermissions(role Role) (perms []model.Permission, err error) {
	permMap := make(map[uint]bool)
	for _, resource := range role.Resources {
		for _, verb := range resource.Verbs {
			scope := resource.Name + ":" + verb
			permID, found := d.permByScope[scope]
			if !found {
				err = &ScopeNotFound{scope}
				return
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
