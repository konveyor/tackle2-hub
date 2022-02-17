package settings

import (
	"net/url"
	"os"
)

const (
	EnvAddonSecretPath = "ADDON_SECRET_PATH"
	EnvWorkingDirPath  = "ADDON_WORKINGDIR_PATH"
	EnvHubBaseURL      = "HUB_BASE_URL"
)

//
// Addon settings.
type Addon struct {
	// Hub settings.
	Hub struct {
		// URL for the hub API.
		URL string
	}
	// Path.
	Path struct {
		// Working directory path.
		WorkingDir string
		// Secret path.
		Secret string
	}
}

func (r *Addon) Load() (err error) {
	var found bool
	r.Hub.URL, found = os.LookupEnv(EnvHubBaseURL)
	if !found {
		r.Hub.URL = "http://localhost:8080"
	}
	_, err = url.Parse(r.Hub.URL)
	if err != nil {
		panic(err)
	}
	r.Path.Secret, found = os.LookupEnv(EnvAddonSecretPath)
	if !found {
		r.Path.Secret = "/tmp/secret.json"
	}
	r.Path.WorkingDir, found = os.LookupEnv(EnvWorkingDirPath)
	if !found {
		r.Path.WorkingDir = "/tmp"
	}

	return
}
