package model

type ImportTag struct {
	Model
	Name     string
	TagType  string
	ImportID uint `gorm:"index"`
	Import   *Import
}
