package stakeholdergroup

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Mgmt = api.StakeholderGroup{
		Name:        "Mgmt",
		Description: "Management stakeholder group.",
	}
	Engineering = api.StakeholderGroup{
		Name:        "Engineering",
		Description: "Engineering team.",
	}
	Samples = []api.StakeholderGroup{Mgmt, Engineering}
)
