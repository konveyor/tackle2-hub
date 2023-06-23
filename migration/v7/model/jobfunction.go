package model

type JobFunction struct {
	Model
	UUID         string `gorm:"uniqueIndex"`
	Username     string
	Name         string        `gorm:"index;unique;not null"`
	Stakeholders []Stakeholder `gorm:"constraint:OnDelete:SET NULL"`
}
