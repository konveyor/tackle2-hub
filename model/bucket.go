package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	path "path"
)

type BucketOwner struct {
	Bucket string `gorm:"<-:create"`
}

func (m *BucketOwner) BeforeCreate(db *gorm.DB) (err error) {
	err = m.Create()
	return
}

func (m *BucketOwner) BeforeDelete(db *gorm.DB) (err error) {
	err = m.Delete()
	return
}

//
// Create associated storage.
func (m *BucketOwner) Create() (err error) {
	uid := uuid.New()
	m.Bucket = path.Join(
		Settings.Hub.Bucket.Path,
		uid.String())
	err = os.MkdirAll(m.Bucket, 0777)
	if err != nil {
		return
	}
	return
}

//
// Purge associated storage.
func (m *BucketOwner) Purge() (err error) {
	content, _ := ioutil.ReadDir(m.Bucket)
	for _, n := range content {
		p := path.Join(m.Bucket, n.Name())
		_ = os.RemoveAll(p)
	}
	return
}

//
// Delete associated storage.
func (m *BucketOwner) Delete() (err error) {
	err = os.RemoveAll(m.Bucket)
	return
}
