package v24

import (
	liberr "github.com/jortel/go-utils/error"
	v23 "github.com/konveyor/tackle2-hub/internal/migration/v23/model"
	"github.com/konveyor/tackle2-hub/internal/migration/v24/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	err = r.migrateIdpRefreshTokens(db)
	if err != nil {
		return
	}
	err = r.setGrantFKs(db)
	if err != nil {
		return
	}
	err = r.dropIdpIdentityColumns(db)
	return
}

func (r Migration) Models() []any {
	return model.All()
}

func (r Migration) migrateIdpRefreshTokens(db *gorm.DB) (err error) {
	var identities []*v23.IdpIdentity
	err = db.Find(&identities).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, identity := range identities {
		if identity.RefreshToken == "" {
			continue
		}
		var grants []*model.Grant
		err = db.Find(&grants, "subject", identity.Subject).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		for _, grant := range grants {
			grant.IdpRefreshToken = identity.RefreshToken
			err = db.Save(grant).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}

func (r Migration) setGrantFKs(db *gorm.DB) (err error) {
	var grants []*model.Grant
	err = db.Find(&grants).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, grant := range grants {
		var user model.User
		err = db.First(&user, "subject", grant.Subject).Error
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
		err = db.First(&identity, "subject", grant.Subject).Error
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
		err = db.First(&client, "subject", grant.Subject).Error
		if err == nil {
			grant.IdpClientID = &client.ID
			err = db.Save(grant).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			continue
		}
	}
	return
}

func (r Migration) dropIdpIdentityColumns(db *gorm.DB) (err error) {
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
	err = migrator.DropColumn(&model.IdpIdentity{}, "LastAuthenticated")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(&model.IdpIdentity{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
