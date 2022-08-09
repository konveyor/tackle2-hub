package model

import (
	"github.com/konveyor/tackle2-hub/model"
)

//
// Proxy configuration.
// kind = (http|https)
type Proxy struct {
	Model
	Enabled    bool
	Kind       string `gorm:"uniqueIndex"`
	Host       string `gorm:"not null"`
	Port       int
	Excluded   model.JSON `json:"excluded"`
	IdentityID *uint      `gorm:"index"`
	Identity   *Identity
}
