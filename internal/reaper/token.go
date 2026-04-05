package reaper

import "gorm.io/gorm"

// TokenReaper reaps api keys.
type TokenReaper struct {
	// DB
	DB *gorm.DB
}

func (r *TokenReaper) Run() {
	Log.V(1).Info("Reaping tokens.")

}
