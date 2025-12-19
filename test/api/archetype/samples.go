package archetype

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	MinimalArchetype = api.Archetype{
		Name:        "Minimal",
		Description: "Archetype minimal sample 1",
		Comments:    "Archetype comments",
	}
	WithProfiles = api.Archetype{
		Name:        "WithProfiles",
		Description: "Archetype with profiles",
		Comments:    "Archetype comments",
		Profiles: []api.TargetProfile{
			{Name: "openshift"},
		},
	}
	Samples = []api.Archetype{MinimalArchetype, WithProfiles}
)
