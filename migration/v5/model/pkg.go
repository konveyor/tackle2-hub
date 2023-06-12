package model

import (
	v4 "github.com/konveyor/tackle2-hub/migration/v4/model"
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
type Model = v4.Model
type Bucket = v4.Bucket
type BucketOwner = v4.BucketOwner
type Dependency = v4.Dependency
type File = v4.File
type Import = v4.Import
type ImportSummary = v4.ImportSummary
type ImportTag = v4.ImportTag
type Proxy = v4.Proxy
type Review = v4.Review
type Setting = v4.Setting
type Tag = v4.Tag
type TagCategory = v4.TagCategory
type Task = v4.Task
type TaskGroup = v4.TaskGroup
type TaskReport = v4.TaskReport
type TTL = v4.TTL

//
// Errors
type DependencyCyclicError = v4.DependencyCyclicError

//
// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []interface{} {
	return []interface{}{
		TechDependency{},
		Incident{},
		Issue{},
		Analysis{},
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
		RuleSet{},
		Rule{},
		MigrationWave{},
	}
}
