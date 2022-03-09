package model

import (
	"gorm.io/gorm"
	"os"
)

type Bucket struct {
	Model
	Name          string `gorm:"uniqueIndex:A"`
	Path          string
	ApplicationID uint `gorm:"uniqueIndex:A"`
	Application   Application
}

func (m *Bucket) AfterDelete(db *gorm.DB) (err error) {
	err = os.RemoveAll(m.Path)
	return
}
