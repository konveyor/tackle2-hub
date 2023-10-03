package model

import (
	"fmt"
	"gorm.io/gorm"
	"sync"
	"time"
)

type Application struct {
	Model
	BucketOwner
	Name              string `gorm:"index;unique;not null"`
	Description       string
	Review            *Review `gorm:"constraint:OnDelete:CASCADE"`
	Repository        JSON    `gorm:"type:json"`
	Binary            string
	Facts             []Fact `gorm:"constraint:OnDelete:CASCADE"`
	Comments          string
	Tasks             []Task     `gorm:"constraint:OnDelete:CASCADE"`
	Tags              []Tag      `gorm:"many2many:ApplicationTags"`
	Identities        []Identity `gorm:"many2many:ApplicationIdentity;constraint:OnDelete:CASCADE"`
	BusinessServiceID *uint      `gorm:"index"`
	BusinessService   *BusinessService
	OwnerID           *uint         `gorm:"index"`
	Owner             *Stakeholder  `gorm:"foreignKey:OwnerID"`
	Contributors      []Stakeholder `gorm:"many2many:ApplicationContributors;constraint:OnDelete:CASCADE"`
	Analyses          []Analysis    `gorm:"constraint:OnDelete:CASCADE"`
	MigrationWaveID   *uint         `gorm:"index"`
	MigrationWave     *MigrationWave
	Ticket            *Ticket `gorm:"constraint:OnDelete:CASCADE"`
}

type Fact struct {
	ApplicationID uint   `gorm:"<-:create;primaryKey"`
	Key           string `gorm:"<-:create;primaryKey"`
	Source        string `gorm:"<-:create;primaryKey;not null"`
	Value         JSON   `gorm:"type:json;not null"`
	Application   *Application
}

//
// ApplicationTag represents a row in the join table for the
// many-to-many relationship between Applications and Tags.
type ApplicationTag struct {
	ApplicationID uint        `gorm:"primaryKey"`
	TagID         uint        `gorm:"primaryKey"`
	Source        string      `gorm:"primaryKey;not null"`
	Application   Application `gorm:"constraint:OnDelete:CASCADE"`
	Tag           Tag         `gorm:"constraint:OnDelete:CASCADE"`
}

//
// TableName must return "ApplicationTags" to ensure compatibility
// with the autogenerated join table name.
func (ApplicationTag) TableName() string {
	return "ApplicationTags"
}

//
// depMutex ensures Dependency.Create() is not executed concurrently.
var depMutex sync.Mutex

type Dependency struct {
	Model
	ToID   uint         `gorm:"index"`
	To     *Application `gorm:"foreignKey:ToID;constraint:OnDelete:CASCADE"`
	FromID uint         `gorm:"index"`
	From   *Application `gorm:"foreignKey:FromID;constraint:OnDelete:CASCADE"`
}

//
// Create a dependency synchronized using a mutex.
func (r *Dependency) Create(db *gorm.DB) (err error) {
	depMutex.Lock()
	defer depMutex.Unlock()
	err = db.Create(r).Error
	return
}

//
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

//
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
	Owns             []Application   `gorm:"foreignKey:OwnerID;constraint:OnDelete:SET NULL"`
	Contributes      []Application   `gorm:"many2many:ApplicationContributors;constraint:OnDelete:CASCADE"`
	MigrationWaves   []MigrationWave `gorm:"many2many:MigrationWaveStakeholders;constraint:OnDelete:CASCADE"`
}

type StakeholderGroup struct {
	Model
	Name           string `gorm:"index;unique;not null"`
	Username       string
	Description    string
	Stakeholders   []Stakeholder   `gorm:"many2many:StakeholderGroupStakeholder;constraint:OnDelete:CASCADE"`
	MigrationWaves []MigrationWave `gorm:"many2many:MigrationWaveStakeholderGroups;constraint:OnDelete:CASCADE"`
}

type MigrationWave struct {
	Model
	Name              string             `gorm:"uniqueIndex:MigrationWaveA"`
	StartDate         time.Time          `gorm:"uniqueIndex:MigrationWaveA"`
	EndDate           time.Time          `gorm:"uniqueIndex:MigrationWaveA"`
	Applications      []Application      `gorm:"constraint:OnDelete:SET NULL"`
	Stakeholders      []Stakeholder      `gorm:"many2many:MigrationWaveStakeholders;constraint:OnDelete:CASCADE"`
	StakeholderGroups []StakeholderGroup `gorm:"many2many:MigrationWaveStakeholderGroups;constraint:OnDelete:CASCADE"`
}

type Tag struct {
	Model
	Name       string `gorm:"uniqueIndex:tagA;not null"`
	Username   string
	CategoryID uint `gorm:"uniqueIndex:tagA;index;not null"`
	Category   TagCategory
}

type TagCategory struct {
	Model
	Name     string `gorm:"index;unique;not null"`
	Username string
	Rank     uint
	Color    string
	Tags     []Tag `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

type Ticket struct {
	Model
	// Kind of ticket in the external tracker.
	Kind string `gorm:"not null"`
	// Parent resource that this ticket should belong to in the tracker. (e.g. Jira project)
	Parent string `gorm:"not null"`
	// Custom fields to send to the tracker when creating the ticket
	Fields JSON `gorm:"type:json"`
	// Whether the last attempt to do something with the ticket reported an error
	Error bool
	// Error message, if any
	Message string
	// Whether the ticket was created in the external tracker
	Created bool
	// Reference id in external tracker
	Reference string
	// URL to ticket in external tracker
	Link string
	// Status of ticket in external tracker
	Status        string
	LastUpdated   time.Time
	Application   *Application
	ApplicationID uint `gorm:"uniqueIndex:ticketA;not null"`
	Tracker       *Tracker
	TrackerID     uint `gorm:"uniqueIndex:ticketA;not null"`
}

type Tracker struct {
	Model
	Name        string `gorm:"index;unique;not null"`
	URL         string
	Kind        string
	Identity    *Identity
	IdentityID  uint
	Connected   bool
	LastUpdated time.Time
	Message     string
	Insecure    bool
	Tickets     []Ticket
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
		m[fmt.Sprintf("category%v", i+1)] = tag.Category
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
	Category string
	ImportID uint `gorm:"index"`
	Import   *Import
}
