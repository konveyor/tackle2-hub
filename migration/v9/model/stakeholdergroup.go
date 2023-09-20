package model

type StakeholderGroup struct {
	Model
	Name           string `gorm:"index;unique;not null"`
	Username       string
	Description    string
	Stakeholders   []Stakeholder   `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
	MigrationWaves []MigrationWave `gorm:"many2many:MigrationWaveStakeholderGroups;constraint:OnDelete:CASCADE"`
	Assessments    []Assessment    `gorm:"many2many:AssessmentStakeholderGroups;constraint:OnDelete:CASCADE"`
	Archetypes     []Archetype     `gorm:"many2many:ArchetypeStakeholderGroups;constraint:OnDelete:CASCADE"`
}
