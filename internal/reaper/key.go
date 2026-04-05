package reaper

import "gorm.io/gorm"

// KeyReaper reaps api keys.
type KeyReaper struct {
	// DB
	DB *gorm.DB
}

func (r *KeyReaper) Run() {
	Log.V(1).Info("Reaping API keys.")

}
