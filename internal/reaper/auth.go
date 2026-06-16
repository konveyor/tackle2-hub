package reaper

import (
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
)

// TokenReaper deletes expired tokens.
type TokenReaper struct {
	// DB
	DB *gorm.DB
}

// Run delete API key tokens that have been expired for more than 1 hour.
// Access tokens are CASCADE deleted when their grant is reaped.
func (r *TokenReaper) Run() {
	Log.V(1).Info("Reaping API key tokens.")
	mark := time.Now().Add(-time.Hour)
	err := r.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			var list []*model.Token
			err = tx.Find(&list, "kind = ? AND expiration < ?", auth.KindAPIKey, mark).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			for _, token := range list {
				err = tx.Delete(token).Error
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				auth.IdP.Cache().TokenDeleted(token.ID)
				Log.Info(
					"Expired API key token deleted.",
					"id",
					token.ID)
			}
			return
		})
	if err != nil {
		Log.Error(err, "")
	}
}

// GrantReaper deletes expired grants.
type GrantReaper struct {
	// DB
	DB *gorm.DB
}

// Run delete grants that have been expired for more than 1 hour.
// The delay prevents deleting a token in use by auth.
func (r *GrantReaper) Run() {
	Log.V(1).Info("Reaping grants.")

	mark := time.Now().Add(-time.Hour)
	err := r.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			var list []*model.Grant
			err = tx.Find(&list, "expiration < ?", mark).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			for _, grant := range list {
				err = tx.Delete(grant).Error
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				auth.IdP.Cache().GrantDeleted(grant.ID)
				Log.Info(
					"Expired grant deleted.",
					"id",
					grant.ID)
			}
			return
		})
	if err != nil {
		Log.Error(err, "")
	}
}
