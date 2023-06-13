package model

type Stakeholder struct {
	Model
	Name             string             `gorm:"not null;"`
	Email            string             `gorm:"index;unique;not null"`
	Groups           []StakeholderGroup `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
	BusinessServices []BusinessService  `gorm:"constraint:OnDelete:SET NULL"`
	JobFunctionID    *uint              `gorm:"index"`
	JobFunction      *JobFunction
	Owns             []Application   `gorm:"foreignKey:OwnerID;constraint:OnDelete:SET NULL"`
	Contributes      []Application   `gorm:"many2many:ApplicationContributors;constraint:OnDelete:CASCADE"`
	MigrationWaves   []MigrationWave `gorm:"many2many:MigrationWaveStakeholders;constraint:OnDelete:CASCADE"`
}
