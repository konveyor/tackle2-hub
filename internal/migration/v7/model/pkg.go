package model

import (
	"github.com/konveyor/tackle2-hub/shared/settings"
)

// JSON field (data) type.
type JSON = []byte

var (
	Settings = &settings.Settings
)

// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []any {
	return []any{
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
