package model

type Tag struct {
	Model
	UUID       *string `gorm:"uniqueIndex"`
	Name       string  `gorm:"uniqueIndex:tagA;not null"`
	Username   string
	CategoryID uint `gorm:"uniqueIndex:tagA;index;not null"`
	Category   TagCategory
}
