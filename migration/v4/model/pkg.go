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
type File = v3.File
type Import = v3.Import
type ImportSummary = v3.ImportSummary
type ImportTag = v3.ImportTag
type Proxy = v3.Proxy
type Review = v3.Review
type Setting = v3.Setting
type Tag = v3.Tag
type TagCategory = v3.TagCategory
type Task = v3.Task
type TaskGroup = v3.TaskGroup
type TaskReport = v3.TaskReport
type TTL = v3.TTL

//
// Errors
type DependencyCyclicError = v3.DependencyCyclicError

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
