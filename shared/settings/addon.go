package settings

import (
	"net/url"

	"github.com/konveyor/tackle2-hub/shared/env"
)

const (
	EnvHubBaseURL   = "HUB_BASE_URL"
	EnvHubToken     = "TOKEN"
	EnvTask         = "TASK"
	EnvAddonHomeDir = "ADDON_HOME"
	EnvSharedDir    = "SHARED_PATH"
	EnvCacheDir     = "CACHE_PATH"
)

// Addon settings.
type Addon struct {
	// HomeDir working directory.
	HomeDir string
	// SharedDir shared mount directory.
	SharedDir string
	// CacheDir cache mount directory.
	CacheDir string
	// Task current task id.
	Task int
	// Hub settings.
	Hub struct {
		// URL for the hub API.
		URL string
		// Token for the hub API.
		Token string
	}
}

func (r *Addon) Load() (err error) {
	r.HomeDir = env.Get(EnvAddonHomeDir, "/addon")
	r.SharedDir = env.Get(EnvSharedDir, "/shared")
	r.CacheDir = env.Get(EnvCacheDir, "/cache")
	r.Hub.URL = env.Get(EnvHubBaseURL, "http://localhost:8080")
	r.Hub.Token = env.Get(EnvHubToken, "")
	r.Task = env.GetInt(EnvTask, 0)
	_, err = url.Parse(r.Hub.URL)
	return
}
