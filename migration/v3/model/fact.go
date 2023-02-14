package model

type Fact struct {
	ApplicationID uint         `gorm:"<-:create;primaryKey"`
	Key           string       `gorm:"<-:create;primaryKey"`
	Value         JSON         `gorm:"not null"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE"`
}
