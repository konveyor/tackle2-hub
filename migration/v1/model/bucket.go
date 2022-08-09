package model

type BucketOwner struct {
	Bucket string `gorm:"index"`
}
