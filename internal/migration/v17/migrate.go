package v17

import (
	liberr "github.com/jortel/go-utils/error"
	v16 "github.com/konveyor/tackle2-hub/internal/migration/v16/model"
	"github.com/konveyor/tackle2-hub/internal/migration/v17/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	err = r.updateEncryption(db)
	if err != nil {
		return
	}
	return
}

func (r Migration) Models() []any {
	return model.All()
}

// updateEncryption re-encrypts identities.
// From: AES (deprecated) to: AES+GCM.
// Stronger and fails when the wrong key is used.
func (r Migration) updateEncryption(db *gorm.DB) (err error) {
	var list []v16.Identity
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list {
		m := &list[i]
		err = m.Decrypt()
		if err != nil {
			return
		}
		updated := &model.Identity{
			Password: m.Password,
			Key:      m.Key,
			Settings: m.Settings,
		}
		err = secret.Encrypt(updated)
		if err != nil {
			return
		}
		m.Password = updated.Password
		m.Key = updated.Key
		m.Settings = updated.Settings
		err = db.Save(m).Error
		if err != nil {
			return
		}
	}
	return
}
