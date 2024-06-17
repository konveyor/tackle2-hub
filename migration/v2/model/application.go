package model

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type Application struct {
	Model
	BucketOwner
	Name              string `gorm:"index;unique;not null"`
	Description       string
	Review            *Review `gorm:"constraint:OnDelete:CASCADE"`
	Repository        JSON    `gorm:"type:json"`
	Binary            string
	Comments          string
	Facts             JSON       `gorm:"type:json"`
	Tasks             []Task     `gorm:"constraint:OnDelete:CASCADE"`
	Tags              []Tag      `gorm:"many2many:ApplicationTags;constraint:OnDelete:CASCADE"`
	Identities        []Identity `gorm:"many2many:ApplicationIdentity;constraint:OnDelete:CASCADE"`
	BusinessServiceID *uint      `gorm:"index"`
	BusinessService   *BusinessService
}

// depMutex ensures Dependency.Create() is not executed concurrently.
var depMutex sync.Mutex

type Dependency struct {
	Model
	ToID   uint         `gorm:"index"`
	To     *Application `gorm:"foreignKey:ToID;constraint:OnDelete:CASCADE"`
	FromID uint         `gorm:"index"`
	From   *Application `gorm:"foreignKey:FromID;constraint:OnDelete:CASCADE"`
}

// Create a dependency synchronized using a mutex.
func (r *Dependency) Create(db *gorm.DB) (err error) {
	depMutex.Lock()
	defer depMutex.Unlock()
	err = db.Create(r).Error
	return
}

// Validation Hook to avoid cyclic dependencies.
func (r *Dependency) BeforeCreate(db *gorm.DB) (err error) {
	var nextDeps []*Dependency
	var nextAppsIDs []uint
	nextAppsIDs = append(nextAppsIDs, r.FromID)
	for len(nextAppsIDs) != 0 {
		db.Where("ToID IN ?", nextAppsIDs).Find(&nextDeps)
		nextAppsIDs = nextAppsIDs[:0] // empty array, but keep capacity
		for _, nextDep := range nextDeps {
			if nextDep.FromID == r.ToID {
				err = DependencyCyclicError{}
				return
			}
			nextAppsIDs = append(nextAppsIDs, nextDep.FromID)
		}
	}

	return
}

// Custom error type to allow API recognize Cyclic Dependency error and assign proper status code.
type DependencyCyclicError struct{}

func (err DependencyCyclicError) Error() string {
	return "cyclic dependencies are not allowed"
}

type BusinessService struct {
	Model
	Name          string `gorm:"index;unique;not null"`
	Description   string
	Applications  []Application `gorm:"constraint:OnDelete:SET NULL"`
	StakeholderID *uint         `gorm:"index"`
	Stakeholder   *Stakeholder
}

type JobFunction struct {
	Model
	Username     string
	Name         string        `gorm:"index;unique;not null"`
	Stakeholders []Stakeholder `gorm:"constraint:OnDelete:SET NULL"`
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

type StakeholderGroup struct {
	Model
	Name         string `gorm:"index;unique;not null"`
	Username     string
	Description  string
	Stakeholders []Stakeholder `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
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

func (r *Import) AsMap() (m map[string]any) {
	m = make(map[string]any)
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
