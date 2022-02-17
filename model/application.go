package model

import "fmt"

type Application struct {
	Model
	Name              string `gorm:"index;unique;not null"`
	Description       string
	Review            *Review
	Repository        JSON
	Comments          string
	Tags              []Tag      `gorm:"many2many:applicationTags"`
	Identities        []Identity `gorm:"many2many:appIdentity"`
	BusinessServiceID uint       `gorm:"index"`
	BusinessService   *BusinessService
}

type Dependency struct {
	Model
	ToID   uint         `gorm:"index"`
	To     *Application `gorm:"foreignKey:ToID;constraint:OnDelete:CASCADE"`
	FromID uint         `gorm:"index"`
	From   *Application `gorm:"foreignKey:FromID;constraint:OnDelete:CASCADE"`
}

type Review struct {
	Model
	BusinessCriticality uint   `gorm:"not null"`
	EffortEstimate      string `gorm:"not null"`
	ProposedAction      string `gorm:"not null"`
	WorkPriority        uint   `gorm:"not null"`
	Comments            string
	Application         *Application
	ApplicationID       uint `gorm:"uniqueIndex"`
}

type Import struct {
	Model
	Filename            string
	ApplicationName     string
	BusinessService     string
	Comments            string
	Dependency          string
	DependencyDirection string
	Description         string
	ErrorMessage        string
	IsValid             bool
	RecordType1         string
	ImportSummary       ImportSummary
	ImportSummaryID     uint `gorm:"index"`
	Processed           bool
	ImportTags          []ImportTag `gorm:"constraint:OnDelete:CASCADE"`
}

func (r *Import) AsMap() (m map[string]interface{}) {
	m = make(map[string]interface{})
	m["filename"] = r.Filename
	m["applicationName"] = r.ApplicationName
	// "Application Name" is necessary in order for
	// the UI to display the error report correctly.
	m["Application Name"] = r.ApplicationName
	m["businessService"] = r.BusinessService
	m["comments"] = r.Comments
	m["dependency"] = r.Dependency
	m["dependencyDirection"] = r.DependencyDirection
	m["description"] = r.Description
	m["errorMessage"] = r.ErrorMessage
	m["isValid"] = r.IsValid
	m["processed"] = r.Processed
	m["recordType1"] = r.RecordType1
	for i, tag := range r.ImportTags {
		m[fmt.Sprintf("tagType%v", i+1)] = tag.TagType
		m[fmt.Sprintf("tag%v", i+1)] = tag.Name
	}
	return
}

type ImportSummary struct {
	Model
	Content      []byte
	Filename     string
	ImportStatus string
	Imports      []Import `gorm:"constraint:OnDelete:CASCADE"`
}

type ImportTag struct {
	Model
	Name     string
	TagType  string
	ImportID uint `gorm:"index"`
	Import   *Import
}
