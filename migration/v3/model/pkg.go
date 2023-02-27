package model

import (
	v2 "github.com/konveyor/tackle2-hub/migration/v2/model"
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
type Model = v2.Model
type BusinessService = v2.BusinessService
type Dependency = v2.Dependency
type Identity = v2.Identity
type JobFunction = v2.JobFunction
type Proxy = v2.Proxy
type Review = v2.Review
type Setting = v2.Setting
type Stakeholder = v2.Stakeholder
type StakeholderGroup = v2.StakeholderGroup
type TagType = v2.TagType
type TaskReport = v2.TaskReport

//
// Errors
type DependencyCyclicError = v2.DependencyCyclicError

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
	}
}
