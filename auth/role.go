package auth

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/settings"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

var Settings = &settings.Settings

type Role struct {
	Name      string     `yaml:"role"`
	Users     []string   `yaml:"users"`
	Resources []Resource `yaml:"resources"`
}

type Resource struct {
	Name  string   `yaml:"name"`
	Verbs []string `yaml:"verbs"`
}

func LoadRoles(path string) (roles []Role, err error) {
	roleFile, err := os.Open(path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer roleFile.Close()

	yamlBytes, err := ioutil.ReadAll(roleFile)
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
