package settings

import "net/url"

const (
	EnvHubBaseURL   = "HUB_BASE_URL"
	EnvHubToken     = "TOKEN"
	EnvTask         = "TASK"
	EnvAddonHomeDir = "ADDON_HOME"
	EnvSharedPath   = "SHARED_PATH"
	EnvCachePath    = "CACHE_PATH"
)

// Addon settings.
type Addon struct {
	// HomeDir working directory.
	HomeDir   string
	SharedDir string
	CacheDir  string
	// Hub settings.
	Hub struct {
		// URL for the hub API.
		URL string
		// Token for the hub API.
		Token string
	}
	//
	Task int
}

func (r *Addon) Load() (err error) {
	r.HomeDir = GetString(EnvAddonHomeDir, "/addon")
	r.SharedDir = GetString(EnvSharedPath, "/shared")
	r.CacheDir = GetString(EnvCachePath, "/cache")
	r.Hub.URL = GetString(EnvHubBaseURL, "http://localhost:8080")
	r.Hub.Token = GetString(EnvHubToken, "")
	r.Task = GetInt(EnvTask, 0)
	_, err = url.Parse(r.Hub.URL)
	return
}
