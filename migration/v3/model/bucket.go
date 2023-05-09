package model

import (
	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"gorm.io/gorm"
	"os"
	"path"
	"time"
)

type Bucket struct {
	Model
	Path       string `gorm:"<-:create;uniqueIndex"`
	Expiration *time.Time
}

func (m *Bucket) BeforeCreate(db *gorm.DB) (err error) {
	if m.Path == "" {
		uid := uuid.New()
		m.Path = path.Join(
			Settings.Hub.Bucket.Path,
			uid.String())
		err = os.MkdirAll(m.Path, 0777)
		if err != nil {
			err = liberr.Wrap(
				err,
				"path",
				m.Path)
		}
	}
	return
}

type BucketOwner struct {
	BucketID *uint `gorm:"index" ref:"bucket"`
	Bucket   *Bucket
}

func (m *BucketOwner) BeforeCreate(db *gorm.DB) (err error) {
	if !m.HasBucket() {
		b := &Bucket{}
		err = db.Create(b).Error
		m.SetBucket(&b.ID)
	}
	return
}

func (m *BucketOwner) SetBucket(id *uint) {
	m.BucketID = id
	m.Bucket = nil
}

func (m *BucketOwner) HasBucket() (b bool) {
	return m.BucketID != nil
}
