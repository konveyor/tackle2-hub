package model

type TagCategory struct {
	Model
	Name     string `gorm:"index;unique;not null"`
	Username string
	Rank     uint
	Color    string
	Tags     []Tag `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}
