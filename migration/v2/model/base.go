package model

import (
	"time"
)

//
// Model Base model.
type Model struct {
	ID         uint      `gorm:"<-:create;primaryKey"`
	CreateTime time.Time `gorm:"<-:create;autoCreateTime"`
	CreateUser string    `gorm:"<-:create"`
	UpdateUser string
}
