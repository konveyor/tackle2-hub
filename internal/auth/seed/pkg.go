package seed

// Role use to read roles.yaml.
type Role struct {
	ID        uint   `yaml:"id"`
	Name      string `yaml:"role"`
	Resources []struct {
		Name  string   `yaml:"name"`
		Verbs []string `yaml:"verbs"`
	} `yaml:"resources"`
}

// User used to read users.yaml.
type User struct {
	ID       uint     `yaml:"id"`
	Userid   string   `yaml:"userid"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`
}
