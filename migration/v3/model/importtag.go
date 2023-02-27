package model

type ImportTag struct {
	Model
	Name     string
	Category string
	ImportID uint `gorm:"index"`
	Import   *Import
}
