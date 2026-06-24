package v24

import (
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/migration/v24/model"
	"gorm.io/gorm"
)

var Log = logr.WithName("migration|v24")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	// AutoMigrate first to add new columns
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	// Migrate data into new columns
	err = r.migrateIdpRefreshTokens(db)
	if err != nil {
		return
	}
	// Drop old columns from IdpIdentity
	err = r.dropIdpIdentityColumns(db)
	return
}

func (r Migration) Models() []any {
	return model.All()
}

func (r Migration) migrateIdpRefreshTokens(db *gorm.DB) (err error) {
	Log.Info("Migrating IdpIdentity refresh tokens to Grant.")

	// Temporary struct to read old IdpIdentity columns that still exist
	type OldIdpIdentity struct {
		model.IdpIdentity
		RefreshToken string `gorm:"column:RefreshToken"`
		Scopes       string `gorm:"column:Scopes"`
	}

	// Fetch all IdpIdentities with old columns
	var identities []*OldIdpIdentity
	err = db.Table("idpidentities").Find(&identities).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	Log.Info("Found IdpIdentities to migrate.", "count", len(identities))

	// For each IdpIdentity, find all associated Grants and copy the refresh token and scopes
	for _, identity := range identities {
		if identity.RefreshToken == "" {
			continue
		}

		// Fetch v24 grants that match this identity's subject
		var grants []*model.Grant
		err = db.Where("subject = ?", identity.Subject).Find(&grants).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}

		// Parse scopes from space-delimited string to JSON array
		var idpScopes []string
		if identity.Scopes != "" {
			idpScopes = strings.Fields(identity.Scopes)
		}

		// Update each grant with refresh token, scopes, and IdpIdentityID FK
		for _, grant := range grants {
			grant.IdpRefreshToken = identity.RefreshToken
			grant.IdpIdentityID = &identity.ID
			grant.IdpScopes = idpScopes

			err = db.Save(grant).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			Log.V(1).Info(
				"Migrated refresh token and scopes from IdpIdentity to Grant.",
				"identity", identity.ID,
				"grant", grant.ID)
		}
	}

	// Now set all FK references on all grants based on their subject
	err = r.setGrantFKs(db)
	if err != nil {
		return
	}

	Log.Info("IdpIdentity refresh token migration completed.")
	return
}

func (r Migration) setGrantFKs(db *gorm.DB) (err error) {
	Log.Info("Setting Grant foreign keys based on subject.")

	var grants []*model.Grant
	err = db.Find(&grants).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	for _, grant := range grants {
		// Find what type of subject this is
		var user model.User
		err = db.Where("subject = ?", grant.Subject).First(&user).Error
		if err == nil {
			grant.UserID = &user.ID
			err = db.Save(grant).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			continue
		}

		var identity model.IdpIdentity
		err = db.Where("subject = ?", grant.Subject).First(&identity).Error
		if err == nil {
			grant.IdpIdentityID = &identity.ID
			err = db.Save(grant).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			continue
		}

		var client model.IdpClient
		err = db.Where("subject = ?", grant.Subject).First(&client).Error
		if err == nil {
			grant.IdpClientID = &client.ID
			err = db.Save(grant).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			continue
		}

		// No subject found - log but don't fail
		Log.V(1).Info(
			"Grant has no matching subject entity.",
			"grant", grant.ID,
			"subject", grant.Subject)
	}

	Log.Info("Grant foreign keys set.")
	return
}

func (r Migration) dropIdpIdentityColumns(db *gorm.DB) (err error) {
	Log.Info("Dropping old IdpIdentity columns.")

	migrator := db.Migrator()
	err = migrator.DropColumn(&model.IdpIdentity{}, "RefreshToken")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropColumn(&model.IdpIdentity{}, "Expiration")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropColumn(&model.IdpIdentity{}, "LastRefreshed")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropColumn(&model.IdpIdentity{}, "Scopes")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	Log.Info("Dropped old IdpIdentity columns.")
	return
}
