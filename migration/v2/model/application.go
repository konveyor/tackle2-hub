package model

type Application struct {
	Model
	BucketOwner
	Name              string `gorm:"index;unique;not null"`
	Description       string
	Review            *Review `gorm:"constraint:OnDelete:CASCADE"`
	Repository        JSON
	Binary            string
	Facts             JSON
	Comments          string
	Tasks             []Task     `gorm:"constraint:OnDelete:CASCADE"`
	Tags              []Tag      `gorm:"many2many:ApplicationTags;constraint:OnDelete:CASCADE"`
	Identities        []Identity `gorm:"many2many:ApplicationIdentity;constraint:OnDelete:CASCADE"`
	BusinessServiceID *uint      `gorm:"index"`
	BusinessService   *BusinessService
}
