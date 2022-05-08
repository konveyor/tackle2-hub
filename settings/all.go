package settings

import (
	"os"
	"strconv"
)

var Settings TackleSettings

type TackleSettings struct {
	Hub
	Metrics
	Addon
	Auth
}

func (r *TackleSettings) Load() (err error) {
	err = r.Hub.Load()
	if err != nil {
		return
	}
	err = r.Addon.Load()
	if err != nil {
		return
	}
	err = r.Auth.Load()
	if err != nil {
		return
	}
	return
}

//
// Get boolean.
func getEnvBool(name string, def bool) bool {
	boolean := def
	if s, found := os.LookupEnv(name); found {
		parsed, err := strconv.ParseBool(s)
		if err == nil {
			boolean = parsed
		}
	}

	return boolean
}
