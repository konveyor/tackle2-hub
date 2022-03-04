package model

type BusinessService struct {
	Model
	Name        string `gorm:"index;unique;not null"`
	Description string
	OwnerID     *uint `gorm:"index"`
	Owner       *Stakeholder
}

type StakeholderGroup struct {
	Model
	Name         string `gorm:"index;unique;not null"`
	Username     string
	Description  string
	Stakeholders []Stakeholder `gorm:"many2many:sgStakeholder"`
}

type Stakeholder struct {
	Model
	Name             string             `gorm:"not null;"`
	Email            string             `gorm:"index;unique;not null"`
	Groups           []StakeholderGroup `gorm:"many2many:sgStakeholder"`
	BusinessServices []BusinessService  `gorm:"foreignKey:OwnerID"`
	JobFunctionID    *uint              `gorm:"index"`
	JobFunction      *JobFunction
}

type JobFunction struct {
	Model
	Username     string
	Name         string `gorm:"index;unique;not null"`
	Stakeholders []Stakeholder
}

type Tag struct {
	Model
	Name      string `gorm:"uniqueIndex:tag_a;not null"`
	Username  string
	TagTypeID uint `gorm:"uniqueIndex:tag_a;index;not null"`
	TagType   TagType
}

type TagType struct {
	Model
	Name     string `gorm:"index;unique;not null"`
	Username string
	Rank     uint
	Color    string
	Tags     []Tag
}
