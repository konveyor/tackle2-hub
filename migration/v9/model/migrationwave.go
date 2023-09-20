package model

import "time"

type MigrationWave struct {
	Model
	Name              string             `gorm:"uniqueIndex:MigrationWaveA"`
	StartDate         time.Time          `gorm:"uniqueIndex:MigrationWaveA"`
	EndDate           time.Time          `gorm:"uniqueIndex:MigrationWaveA"`
	Applications      []Application      `gorm:"constraint:OnDelete:SET NULL"`
	Stakeholders      []Stakeholder      `gorm:"many2many:MigrationWaveStakeholders;constraint:OnDelete:CASCADE"`
	StakeholderGroups []StakeholderGroup `gorm:"many2many:MigrationWaveStakeholderGroups;constraint:OnDelete:CASCADE"`
}
