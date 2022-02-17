package model

type Setting struct {
	ID    uint   `gorm:"primaryKey"`
	Key   string `gorm:"uniqueIndex"`
	Value JSON
}
