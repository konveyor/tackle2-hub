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
	Login    string   `yaml:"login"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`
}

// IdpClient used to read clients.yaml.
type IdpClient struct {
	ID              uint     `yaml:"id"`
	ClientId        string   `yaml:"clientId"`
	Secret          string   `yaml:"secret"`
	ApplicationType string   `yaml:"applicationType"`
	Grants          []string `yaml:"grants"`
	RedirectURIs    []string `yaml:"redirectURIs"`
	Scopes          []string `yaml:"scopes"`
}
