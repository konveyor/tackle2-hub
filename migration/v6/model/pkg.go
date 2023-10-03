package model

import "github.com/konveyor/tackle2-hub/settings"

var (
	Settings = &settings.Settings
)

//
// JSON field (data) type.
type JSON = []byte

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
