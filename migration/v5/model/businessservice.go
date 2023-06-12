package model

type BusinessService struct {
	Model
	Name          string `gorm:"index;unique;not null"`
	Description   string
	Applications  []Application `gorm:"constraint:OnDelete:SET NULL"`
	StakeholderID *uint         `gorm:"index"`
	Stakeholder   *Stakeholder
}
