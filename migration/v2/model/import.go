package model

import "fmt"

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
