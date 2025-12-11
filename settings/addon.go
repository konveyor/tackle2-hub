package settings

import (
	shared "github.com/konveyor/tackle2-hub/shared/settings"
)

const (
	EnvHubBaseURL   = shared.EnvHubBaseURL
	EnvHubToken     = shared.EnvHubToken
	EnvTask         = shared.EnvTask
	EnvAddonHomeDir = shared.EnvAddonHomeDir
	EnvSharedDir    = shared.EnvSharedDir
	EnvCacheDir     = shared.EnvCacheDir
)

// Addon settings.
type Addon = shared.Addon
