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
	Excluded   JSON  `gorm:"type:json"`
	IdentityID *uint `gorm:"index"`
	Identity   *Identity
}
