package settings

import (
	"os"
	"strconv"
)

// GetString returns an envar.
func GetString(name string, def string) (v string) {
	v, found := os.LookupEnv(name)
	if !found {
		v = def
	}
	return
}

// GetInt returns an envar.
func GetInt(name string, def int) (v int) {
	s, found := os.LookupEnv(name)
	if found {
		v, _ = strconv.Atoi(s)
	} else {
		v = def
	}
	return
}

// GetBool returns an envar.
func GetBool(name string, def bool) (v bool) {
	s, found := os.LookupEnv(name)
	if found {
		v, _ = strconv.ParseBool(s)
	} else {
		v = def
	}
	return
}
