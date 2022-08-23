package model

type Volume struct {
	Model
	Name     string `gorm:"<-:create;unique"`
	Capacity string
	Used     string
}
