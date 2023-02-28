package model

type Fact struct {
	ApplicationID uint   `gorm:"<-:create;primaryKey"`
	Key           string `gorm:"<-:create;primaryKey"`
	Value         JSON   `gorm:"type:json;not null"`
	Application   *Application
}
