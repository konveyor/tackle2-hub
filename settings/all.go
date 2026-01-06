package settings

import (
	"gopkg.in/yaml.v2"
)

var Settings TackleSettings

func init() {
	err := Settings.Load()
	if err != nil {
		panic(err)
	}
}

type TackleSettings struct {
	Addon
}

func (r *TackleSettings) Load() (err error) {
	err = r.Addon.Load()
	if err != nil {
		return
	}
	return
}

func (r TackleSettings) String() (s string) {
	b, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}
	s = string(b)
	return
}
