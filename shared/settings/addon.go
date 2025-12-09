package settings

import (
	"net/url"
	"os"
	"strconv"
)

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
	var found bool
	r.HomeDir, found = os.LookupEnv(EnvAddonHomeDir)
	if !found {
		r.HomeDir = "/addon"
	}
	r.SharedDir, found = os.LookupEnv(EnvSharedPath)
	if !found {
		r.SharedDir = "/shared"
	}
	r.CacheDir, found = os.LookupEnv(EnvCachePath)
	if !found {
		r.CacheDir = "/cache"
	}
	r.Hub.URL, found = os.LookupEnv(EnvHubBaseURL)
	if !found {
		r.Hub.URL = "http://localhost:8080"
	}
	_, err = url.Parse(r.Hub.URL)
	if err != nil {
		panic(err)
	}
	r.Hub.Token, found = os.LookupEnv(EnvHubToken)
	if s, found := os.LookupEnv(EnvTask); found {
		r.Task, _ = strconv.Atoi(s)
	}

	return
}
