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
)

// Addon settings.
type Addon struct {
	// HomeDir working directory.
	HomeDir string
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
