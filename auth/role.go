package auth

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/settings"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

var Settings = &settings.Settings

//
// AddonRole defines the addon scopes.
var AddonRole = []string{
	"applications:get",
	"applications:put",
	"applications.tags:*",
	"applications.bucket:*",
	"identities:get",
	"identities:decrypt",
	"proxies:get",
	"settings:get",
	"tags:*",
	"tagtypes:*",
	"tasks:get",
	"tasks.report:*",
	"tasks.bucket:get",
	"files:*",
}

//
// Role represents a RBAC role which grants
// access to particular resources in the hub.
type Role struct {
	Name      string     `yaml:"role"`
	Resources []Resource `yaml:"resources"`
}

//
// Resource is a set of permissions for a hub resource that a role may have.
type Resource struct {
	Name  string   `yaml:"name"`
	Verbs []string `yaml:"verbs"`
}

//
// User is a hub user which may have Roles.
type User struct {
	// Username
	Name string `yaml:"name"`
	// Default password
	Password string `yaml:"password"`
	// List of roles specified by name
	Roles []string `yaml:"roles"`
}

//
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

	err = yaml.Unmarshal(yamlBytes, &roles)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
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

	err = yaml.Unmarshal(yamlBytes, &users)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
