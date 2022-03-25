package model

type Setting struct {
	Model
	Key   string `gorm:"<-:create;uniqueIndex"`
	Value JSON
}
