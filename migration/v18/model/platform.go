package model

import (
	"github.com/konveyor/tackle2-hub/migration/json"
)

type Manifest struct {
	Model
	Content       json.Map `gorm:"type:json;serializer:json"`
	Secret        json.Map `gorm:"type:json;serializer:json" secret:""`
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
type TargetProfile struct {
	Model
	Name        string      `gorm:"uniqueIndex:targetProfileA;not null"`
	Generators  []Generator `gorm:"many2many:TargetGenerator;constraint:OnDelete:CASCADE"`
	ArchetypeID uint        `gorm:"uniqueIndex:targetProfileA;not null"`
	Archetype   Archetype
}

type Generator struct {
	Model
	Kind       string
	Name       string
	Repository Repository `gorm:"type:json;serializer:json"`
	Params     json.Map   `gorm:"type:json;serializer:json"`
	Values     json.Map   `gorm:"type:json;serializer:json"`
	IdentityID *uint
	Identity   *Identity
	Profiles   []TargetProfile `gorm:"many2many:TargetGenerator;constraint:OnDelete:CASCADE"`
}
