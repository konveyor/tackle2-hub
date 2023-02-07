package model

type TagType struct {
	Model
	Name     string `gorm:"index;unique;not null"`
	Username string
	Rank     uint
	Color    string
	Tags     []Tag `gorm:"constraint:OnDelete:CASCADE"`
}
