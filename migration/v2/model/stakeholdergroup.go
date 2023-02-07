package model

type StakeholderGroup struct {
	Model
	Name         string `gorm:"index;unique;not null"`
	Username     string
	Description  string
	Stakeholders []Stakeholder `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
}
