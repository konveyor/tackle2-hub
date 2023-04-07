package model

type JobFunction struct {
	Model
	Username     string
	Name         string        `gorm:"index;unique;not null"`
	Stakeholders []Stakeholder `gorm:"constraint:OnDelete:SET NULL"`
}
