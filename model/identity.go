package model

import (
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/encryption"
	"gorm.io/gorm"
)

//
// Identity represents and identity with a set of credentials.
type Identity struct {
	Model
	Kind        string `gorm:"not null"`
	Name        string `gorm:"not null"`
	Description string
	User        string
	Password    string
	Key         string
	Settings    string
	Proxies     []Proxy
	Encrypted   string
}

//
// Encrypt sensitive fields.
func (r *Identity) Encrypt() (err error) {
	passphrase := Settings.Encryption.Passphrase
	aes := encryption.New(passphrase)
	encrypted := Identity{}
	encrypted.User = r.User
	encrypted.Password = r.Password
	encrypted.Key = r.Key
	encrypted.Settings = r.Settings
	b, err := json.Marshal(encrypted)
	if err != nil {
		err = liberr.Wrap(
			err,
			"id",
			r.ID)
		return
	}
	r.Encrypted, err = aes.Encrypt(string(b))
	return
}

//
// Decrypt sensitive fields.
func (r *Identity) Decrypt(passphrase string) (err error) {
	aes := encryption.New(passphrase)
	decrypted := &Identity{}
	var dj string
	dj, err = aes.Decrypt(r.Encrypted)
	if err != nil {
		err = liberr.Wrap(
			err,
			"id",
			r.ID)
		return
	}
	err = json.Unmarshal([]byte(dj), decrypted)
	if err != nil {
		err = liberr.Wrap(
			err,
			"id",
			r.ID)
		return
	}
	r.User = decrypted.User
	r.Password = decrypted.Password
	r.Key = decrypted.Key
	r.Settings = decrypted.Settings
	return
}

//
// BeforeSave ensure encrypted.
func (r *Identity) BeforeSave(tx *gorm.DB) (err error) {
	err = r.Encrypt()
	if err == nil {
		r.User = ""
		r.Password = ""
		r.Key = ""
		r.Settings = ""
	}
	return
}

//
// BeforeUpdate ensure encrypted.
func (r *Identity) BeforeUpdate(tx *gorm.DB) (err error) {
	err = r.BeforeSave(tx)
	return
}
