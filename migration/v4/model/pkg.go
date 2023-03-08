package model

import (
	v3 "github.com/konveyor/tackle2-hub/migration/v3/model"
	"github.com/konveyor/tackle2-hub/settings"
)

var (
	Settings = &settings.Settings
)

//
// JSON field (data) type.
type JSON = []byte

//
// Unchanged models imported from previous migration.
type Model = v3.Model
type Bucket = v3.Bucket
type BucketOwner = v3.BucketOwner
type Dependency = v3.Dependency
type Fact = v3.Fact
type File = v3.File
type Identity = v3.Identity
type Import = v3.Import
type ImportSummary = v3.ImportSummary
type ImportTag = v3.ImportTag
type Proxy = v3.Proxy
type Review = v3.Review
type RuleBundle = v3.RuleBundle
type RuleSet = v3.RuleSet
type Setting = v3.Setting
type Tag = v3.Tag
type TagCategory = v3.TagCategory
type Task = v3.Task
type TaskGroup = v3.TaskGroup
type TaskReport = v3.TaskReport
type Ticket = v3.Ticket
type Tracker = v3.Tracker
type TTL = v3.TTL
type Metadata = v3.Metadata
type Project = v3.Project
type IssueType = v3.IssueType

//
// Errors
type DependencyCyclicError = v3.DependencyCyclicError

//
// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []interface{} {
	return []interface{}{
		ImportSummary{},
		Import{},
		ImportTag{},
		JobFunction{},
		TagCategory{},
		Tag{},
		StakeholderGroup{},
		Stakeholder{},
		BusinessService{},
		Bucket{},
		Application{},
		ApplicationTag{},
		Dependency{},
		Review{},
		Identity{},
		Task{},
		TaskGroup{},
		TaskReport{},
		Proxy{},
		Tracker{},
		Ticket{},
		File{},
		Fact{},
		RuleBundle{},
		RuleSet{},
		MigrationWave{},
	}
}
