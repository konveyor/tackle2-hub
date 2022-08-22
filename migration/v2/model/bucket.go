package model

import (
	"github.com/google/uuid"
	liberr "github.com/konveyor/controller/pkg/error"
	"gorm.io/gorm"
	"os"
	"path"
)

type BucketOwner struct {
	Bucket string `gorm:"index"`
}

func (m *BucketOwner) BeforeCreate(db *gorm.DB) (err error) {
	uid := uuid.New()
	m.Bucket = path.Join(
		Settings.Hub.Bucket.Path,
		uid.String())
	err = os.MkdirAll(m.Bucket, 0777)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			m.Bucket)
	}
	return
}
