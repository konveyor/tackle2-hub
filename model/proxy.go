package model

//
// Proxy configuration.
// kind = (http|https)
type Proxy struct {
	Model
	Kind       string `gorm:"uniqueIndex"`
	Host       string `gorm:"not null"`
	Port       int
	IdentityID uint `gorm:"index"`
}
