package model

type BusinessService struct {
	Model
	Name          string `gorm:"index;unique;not null"`
	Description   string
	Applications  []Application `gorm:"constraint:OnDelete:SET NULL"`
	StakeholderID *uint         `gorm:"index"`
	Stakeholder   *Stakeholder
}

type StakeholderGroup struct {
	Model
	Name         string `gorm:"index;unique;not null"`
	Username     string
	Description  string
	Stakeholders []Stakeholder `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
}

type Stakeholder struct {
	Model
	Name             string             `gorm:"not null;"`
	Email            string             `gorm:"index;unique;not null"`
	Groups           []StakeholderGroup `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
	BusinessServices []BusinessService  `gorm:"constraint:OnDelete:SET NULL"`
	JobFunctionID    *uint              `gorm:"index"`
	JobFunction      *JobFunction
}

type JobFunction struct {
	Model
	Username     string
	Name         string        `gorm:"index;unique;not null"`
	Stakeholders []Stakeholder `gorm:"constraint:OnDelete:SET NULL"`
}

type Tag struct {
	Model
	Name      string `gorm:"uniqueIndex:tagA;not null"`
	Username  string
	TagTypeID uint `gorm:"uniqueIndex:tagA;index;not null"`
	TagType   TagType
}

type TagType struct {
	Model
	Name     string `gorm:"index;unique;not null"`
	Username string
	Rank     uint
	Color    string
	Tags     []Tag `gorm:"constraint:OnDelete:CASCADE"`
}
