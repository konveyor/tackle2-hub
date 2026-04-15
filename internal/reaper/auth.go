package reaper

import (
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
)

// KeyReaper deletes expired api-keys.
type KeyReaper struct {
	// DB
	DB *gorm.DB
}

// Run delete api-keys that have been expired for more than 1 hour.
// The delay prevents deleting a token in use by auth.
func (r *KeyReaper) Run() {
	Log.V(1).Info("Reaping API keys.")

	mark := time.Now().Add(-time.Hour)

	err := r.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			var list []*model.APIKey
			err = tx.Find(&list, "expiration < ?", mark).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			for _, key := range list {
				err = tx.Delete(key).Error
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				Log.Info(
					"Expired APIKey deleted.",
					"id",
					key.ID,
					"digest",
					key.Digest)
			}
			return
		})
	if err != nil {
		Log.Error(err, "")
	}
}

// TokenReaper deletes expired tokens.
type TokenReaper struct {
	// DB
	DB *gorm.DB
}

// Run delete tokens that have been expired for more than 1 hour.
// The delay prevents deleting a token in use by auth.
func (r *TokenReaper) Run() {
	Log.V(1).Info("Reaping tokens.")

	mark := time.Now().Add(-time.Hour)

	err := r.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			var list []*model.Token
			err = tx.Find(&list, "expiration < ?", mark).Error
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
				Log.Info(
					"Expired token deleted.",
					"id",
					token.ID,
					"grant",
					token.GrantId)
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
				Log.Info(
					"Expired grant deleted.",
					"id",
					grant.ID,
					"grant",
					grant.GrantId,
					"refreshToken",
					grant.RefreshToken)
			}
			return
		})
	if err != nil {
		Log.Error(err, "")
	}
}
