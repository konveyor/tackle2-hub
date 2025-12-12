package setting

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	SampleSetting = api.Setting{
		Key:   "sample.setting.1",
		Value: "data-123",
	}

	Samples = []api.Setting{SampleSetting}
)
