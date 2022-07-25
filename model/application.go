package model

import "fmt"

type Application struct {
	Model
	BucketOwner
	Name              string `gorm:"index;unique;not null"`
	Description       string
	Review            *Review `gorm:"constraint:OnDelete:CASCADE"`
	Repository        JSON
	Binary            string
	Facts             JSON
	Comments          string
	Tasks             []Task     `gorm:"constraint:OnDelete:CASCADE"`
	Tags              []Tag      `gorm:"many2many:ApplicationTags;constraint:OnDelete:CASCADE"`
	Identities        []Identity `gorm:"many2many:ApplicationIdentity;constraint:OnDelete:CASCADE"`
	BusinessServiceID *uint      `gorm:"index"`
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
	ApplicationID       uint `gorm:"uniqueIndex"`
	Application         *Application
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
	BinaryGroup         string
	BinaryArtifact      string
	BinaryVersion       string
	BinaryPackaging     string
	RepositoryKind      string
	RepositoryURL       string
	RepositoryBranch    string
	RepositoryPath      string
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
	Content        []byte
	Filename       string
	ImportStatus   string
	Imports        []Import `gorm:"constraint:OnDelete:CASCADE"`
	CreateEntities bool
}

type ImportTag struct {
	Model
	Name     string
	TagType  string
	ImportID uint `gorm:"index"`
	Import   *Import
}
