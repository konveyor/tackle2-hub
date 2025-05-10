package model

import "github.com/konveyor/tackle2-hub/migration/json"

type Manifest struct {
	Model
	Content       json.Map `gorm:"type:json;serializer:json"`
	Secret        json.Map `gorm:"type:json;serializer:json" secret:""`
	ApplicationID uint
	Application   Application
}
