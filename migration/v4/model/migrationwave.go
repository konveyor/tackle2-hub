package model

import "time"

type MigrationWave struct {
	Model
	Name              string
	StartDate         time.Time
	EndDate           time.Time
	Applications      []Application      `gorm:"constraint:OnDelete:SET NULL"`
	Stakeholders      []Stakeholder      `gorm:"many2many:MigrationWaveStakeholders;constraint:OnDelete:CASCADE"`
	StakeholderGroups []StakeholderGroup `gorm:"many2many:MigrationWaveStakeholderGroups;constraint:OnDelete:CASCADE"`
}
