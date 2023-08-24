package model

import "github.com/konveyor/tackle2-hub/migration/v6/model"

//
// JSON field (data) type.
type JSON = []byte

type Model = model.Model
type TechDependency = model.TechDependency
type Incident = model.Incident
type Analysis = model.Analysis
type Issue = model.Issue
type Bucket = model.Bucket
type BucketOwner = model.BucketOwner
type Dependency = model.Dependency
type File = model.File
type Fact = model.Fact
type Identity = model.Identity
type Import = model.Import
type ImportSummary = model.ImportSummary
type ImportTag = model.ImportTag
type Proxy = model.Proxy
type Review = model.Review
type Setting = model.Setting
type Task = model.Task
type TaskGroup = model.TaskGroup
type TaskReport = model.TaskReport
type Ticket = model.Ticket
type Tracker = model.Tracker
type TTL = model.TTL
type DependencyCyclicError = model.DependencyCyclicError

//
// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []interface{} {
	return []interface{}{
		Application{},
		TechDependency{},
		Incident{},
		Analysis{},
		Issue{},
		Bucket{},
		BusinessService{},
		Dependency{},
		File{},
		Fact{},
		Identity{},
		Import{},
		ImportSummary{},
		ImportTag{},
		JobFunction{},
		MigrationWave{},
		Proxy{},
		Review{},
		Setting{},
		RuleSet{},
		Rule{},
		Stakeholder{},
		StakeholderGroup{},
		Tag{},
		TagCategory{},
		Task{},
		TaskGroup{},
		TaskReport{},
		Ticket{},
		Tracker{},
		ApplicationTag{},
	}
}
