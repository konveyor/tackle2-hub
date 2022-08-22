package model

//
// Proxy configuration.
// kind = (http|https)
type Proxy struct {
	Model
	Enabled    bool
	Kind       string `gorm:"uniqueIndex"`
	Host       string `gorm:"not null"`
	Port       int
	Excluded   JSON  `json:"excluded"`
	IdentityID *uint `gorm:"index"`
	Identity   *Identity
}
