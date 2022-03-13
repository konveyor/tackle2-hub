package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"os"
	pathlib "path"
)

type Bucket struct {
	Path string `gorm:"<-:create"`
}

func (m *Bucket) BeforeCreate(db *gorm.DB) (err error) {
	uid := uuid.New()
	m.Path = pathlib.Join(
		Settings.Hub.Bucket.Path,
		uid.String())
	err = os.MkdirAll(m.Path, 0777)
	if err != nil {
		return
	}
	return
}

func (m *Bucket) BeforeDelete(db *gorm.DB) (err error) {
	err = os.RemoveAll(m.Path)
	return
}
