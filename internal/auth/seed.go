package auth

import (
	"embed"
	"io/fs"
	"sort"

	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var (
	//go:embed seed
	seedFS  embed.FS
	seedDir fs.FS
)

func init() {
	var err error
	seedDir, err = fs.Sub(seedFS, "seed")
	if err != nil {
		panic(err)
	}
}

type RoleDef struct {
	Name      string `yaml:"role"`
	Resources []struct {
		Name  string   `yaml:"name"`
		Verbs []string `yaml:"verbs"`
	} `yaml:"resources"`
}

type Domain struct {
	db     *gorm.DB
	roles  map[string][]string
	scopes []string
}

func (d *Domain) Load() {
	b, err := fs.ReadFile(seedDir, "roles.yaml")
	if err != nil {
		panic(err)
	}
	rdef := []RoleDef{}
	err = yaml.Unmarshal(b, &rdef)
	if err != nil {
		panic(err)
	}
	d.build(rdef)
}

func (d *Domain) Roles() (roles []string) {
	for r, _ := range d.roles {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	return
}

func (d *Domain) Scopes() (scopes []string) {
	scopes = d.scopes
	return
}

func (d *Domain) FindScopes(roles []string) (scopes []string) {
	mp := make(map[string]byte)
	for _, role := range roles {
		r, found := d.roles[role]
		if !found {
			continue
		}
		for _, s := range r {
			mp[s] = 0
		}
	}
	for s, _ := range mp {
		scopes = append(scopes, s)
	}
	sort.Strings(scopes)
	return
}

func (d *Domain) build(rdef []RoleDef) {
	d.roles = make(map[string][]string)
	scopes := make(map[string]byte)
	for _, role := range rdef {
		for _, resource := range role.Resources {
			for _, verb := range resource.Verbs {
				scope := resource.Name + ":" + verb
				scopes[scope] = 0
				d.roles[role.Name] = append(
					d.roles[role.Name],
					scope)
				scope = "role=" + role.Name
				d.roles[role.Name] = append(
					d.roles[role.Name],
					scope)
			}
		}
	}
	for scope, _ := range scopes {
		d.scopes = append(d.scopes, scope)
	}
	sort.Strings(d.scopes)
	return
}
