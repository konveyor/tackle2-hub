package model

import (
	"github.com/google/uuid"
	liberr "github.com/konveyor/controller/pkg/error"
	"gorm.io/gorm"
	"os"
	"path"
	"time"
)

type File struct {
	Model
	Name       string
	Path       string `gorm:"<-:create;uniqueIndex"`
	Expiration *time.Time
}

func (m *File) BeforeCreate(db *gorm.DB) (err error) {
	uid := uuid.New()
	m.Path = path.Join(
		Settings.Hub.Bucket.Path,
		".file",
		uid.String())
	err = os.MkdirAll(path.Dir(m.Path), 0777)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			m.Path)
	}
	return
}
