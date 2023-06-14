package model

type Application struct {
	Model
	BucketOwner
	Name              string `gorm:"index;unique;not null"`
	Description       string
	Review            *Review `gorm:"constraint:OnDelete:CASCADE"`
	Repository        JSON    `gorm:"type:json"`
	Binary            string
	Facts             []Fact `gorm:"constraint:OnDelete:CASCADE"`
	Comments          string
	Tasks             []Task     `gorm:"constraint:OnDelete:CASCADE"`
	Tags              []Tag      `gorm:"many2many:ApplicationTags"`
	Identities        []Identity `gorm:"many2many:ApplicationIdentity;constraint:OnDelete:CASCADE"`
	BusinessServiceID *uint      `gorm:"index"`
	BusinessService   *BusinessService
	OwnerID           *uint         `gorm:"index"`
	Owner             *Stakeholder  `gorm:"foreignKey:OwnerID"`
	Contributors      []Stakeholder `gorm:"many2many:ApplicationContributors;constraint:OnDelete:CASCADE"`
	Analyses          []Analysis    `gorm:"constraint:OnDelete:CASCADE"`
	MigrationWaveID   *uint         `gorm:"index"`
	MigrationWave     *MigrationWave
}
