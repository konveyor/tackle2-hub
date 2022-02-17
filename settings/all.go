package settings

var Settings TackleSettings

type TackleSettings struct {
	Hub
	Addon
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

	return
}
