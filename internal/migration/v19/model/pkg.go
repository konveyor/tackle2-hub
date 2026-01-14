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
		Application{},
		TechDependency{},
		Incident{},
		Analysis{},
		Insight{},
		Bucket{},
		BusinessService{},
		Dependency{},
		File{},
		Fact{},
		Generator{},
		Identity{},
		Import{},
		ImportSummary{},
		ImportTag{},
		JobFunction{},
		Manifest{},
		MigrationWave{},
		PK{},
		Platform{},
		Proxy{},
		Review{},
		Setting{},
		RuleSet{},
		Rule{},
		Stakeholder{},
		StakeholderGroup{},
		Tag{},
		TagCategory{},
		Target{},
		TargetProfile{},
		Task{},
		TaskGroup{},
		TaskReport{},
		Ticket{},
		Tracker{},
		ApplicationTag{},
		ApplicationIdentity{},
		Questionnaire{},
		Assessment{},
		Archetype{},
		ProfileGenerator{},
	}
}
