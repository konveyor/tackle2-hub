package auth

import (
	"io"
	"os"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gopkg.in/yaml.v2"
)

var Settings = &settings.Settings

// AddonRole defines the addon scopes.
var AddonRole = []string{
	"addons:get",
	"analysis.profiles:get",
	"applications:get",
	"applications:post",
	"applications:put",
	"applications.tags:*",
	"applications.facts:*",
	"applications.bucket:*",
	"applications.analyses:*",
	"applications.manifests:*",
	"archetypes:get",
	"generators:get",
	"identities:get",
	"identities:decrypt",
	"manifests:*",
	"platforms:get",
	"proxies:get",
	"schemas:get",
	"settings:get",
	"tags:*",
	"tagcategories:*",
	"tasks:get",
	"tasks.report:*",
	"tasks.bucket:get",
	"files:*",
	"rulesets:get",
	"targets:get",
}

// Role represents a RBAC role which grants
// access to particular resources in the hub.
type Role struct {
	Name      string     `yaml:"role" validate:"required"`
	Resources []Resource `yaml:"resources" validate:"required"`
}

// Resource is a set of permissions for a hub resource that a role may have.
type Resource struct {
	Name  string   `yaml:"name" validate:"required"`
	Verbs []string `yaml:"verbs" validate:"required,dive,oneof=get post put patch delete"`
}

// User is a hub user which may have Roles.
type User struct {
	// Username
	Name string `yaml:"name"`
	// FirstName
	FirstName string `yaml:"firstName"`
	// LastName
	LastName string `yaml:"lastName"`
	// Email
	Email string `yaml:"email"`
	// Default password
	Password string `yaml:"password"`
	// List of roles specified by name
	Roles []string `yaml:"roles"`
}

// LoadRoles loads a list of Role structs from a yaml file
// that is located at the given path.
func LoadRoles(path string) (roles []Role, err error) {
	roleFile, err := os.Open(path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer roleFile.Close()

	yamlBytes, err := io.ReadAll(roleFile)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = yaml.UnmarshalStrict(yamlBytes, &roles)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// LoadUsers loads a list of User structs from a yaml
// file that is located at the given path.
func LoadUsers(path string) (users []User, err error) {
	userFile, err := os.Open(path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer userFile.Close()

	yamlBytes, err := io.ReadAll(userFile)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = yaml.UnmarshalStrict(yamlBytes, &users)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
