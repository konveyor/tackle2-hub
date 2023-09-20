package model

type Archetype struct {
	Model
	Name              string
	Description       string
	Comments          string
	Review            *Review            `gorm:"constraint:OnDelete:CASCADE"`
	Assessments       []Assessment       `gorm:"constraint:OnDelete:CASCADE"`
	CriteriaTags      []Tag              `gorm:"many2many:ArchetypeCriteriaTags"`
	Tags              []Tag              `gorm:"many2many:ArchetypeTags"`
	Stakeholders      []Stakeholder      `gorm:"many2many:ArchetypeStakeholders;constraint:OnDelete:CASCADE"`
	StakeholderGroups []StakeholderGroup `gorm:"many2many:ArchetypeStakeholderGroups;constraint:OnDelete:CASCADE"`
}
