package model

import (
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
)

// JSON field (data) type.
type JSON = []byte

// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []any {
	return []any{
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
