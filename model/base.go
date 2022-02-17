package model

import (
	"time"
)

//
// Model Base model.
type Model struct {
	ID         uint `gorm:"primaryKey"`
	CreateUser string
	UpdateUser string
	CreateTime time.Time `gorm:"autoCreateTime"`
}
