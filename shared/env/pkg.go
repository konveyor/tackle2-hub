package env

import (
	"os"
	"strconv"
	"time"
)

// Get returns an envar.
func Get(name string, def string) (v string) {
	v, found := os.LookupEnv(name)
	if !found {
		v = def
	}
	return
}

// GetInt returns an envar.
func GetInt(name string, def int) (v int) {
	var err error
	s, found := os.LookupEnv(name)
	if found {
		v, err = strconv.Atoi(s)
		if err != nil {
			v = def
		}
	} else {
		v = def
	}
	return
}

// GetBool returns an envar.
func GetBool(name string, def bool) (v bool) {
	var err error
	s, found := os.LookupEnv(name)
	if found {
		v, err = strconv.ParseBool(s)
		if err != nil {
			v = def
		}
	} else {
		v = def
	}
	return
}

// GetSecond returns an envar.
func GetSecond(name string, def int) (d time.Duration) {
	n := GetInt(name, def)
	d = time.Duration(n) * time.Second
	return
}

// GetMinute returns an envar.
func GetMinute(name string, def int) (d time.Duration) {
	n := GetInt(name, def)
	d = time.Duration(n) * time.Minute
	return
}

// GetHour returns an envar.
func GetHour(name string, def int) (d time.Duration) {
	n := GetInt(name, def)
	d = time.Duration(n) * time.Hour
	return
}

// GetDay returns an envar.
func GetDay(name string, def int) (d time.Duration) {
	n := GetInt(name, def)
	d = time.Duration(n) * time.Hour * 24
	return
}
