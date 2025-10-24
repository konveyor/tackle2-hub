package model

import (
	"github.com/konveyor/tackle2-hub/migration/json"
)

type Manifest struct {
	Model
	Content       json.Map `gorm:"type:jsonb;serializer:json"`
	Secret        json.Map `gorm:"type:jsonb;serializer:json" secret:""`
	ApplicationID uint
	Application   Application
}

type Platform struct {
	Model
	Name         string
	Kind         string
	URL          string
	IdentityID   *uint
	Identity     *Identity
	Applications []Application `gorm:"constraint:OnDelete:SET NULL"`
	Tasks        []Task        `gorm:"constraint:OnDelete:CASCADE"`
}

type Generator struct {
	Model
	UUID        *string `gorm:"uniqueIndex"`
	Kind        string
	Name        string
	Description string
	Repository  Repository `gorm:"type:jsonb;serializer:json"`
	Params      json.Map   `gorm:"type:jsonb;serializer:json"`
	Values      json.Map   `gorm:"type:jsonb;serializer:json"`
	IdentityID  *uint
	Identity    *Identity
	Profiles    []TargetProfile `gorm:"many2many:generator_target_profiles;constraint:OnDelete:CASCADE"`
}
