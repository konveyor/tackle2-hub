package model

type Tag struct {
	Model
	Name      string `gorm:"uniqueIndex:tagA;not null"`
	Username  string
	TagTypeID uint `gorm:"uniqueIndex:tagA;index;not null"`
	TagType   TagType
}
