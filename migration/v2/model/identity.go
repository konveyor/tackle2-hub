package model

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/encryption"
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
	Proxies     []Proxy `gorm:"constraint:OnDelete:SET NULL"`
}

//
// Encrypt sensitive fields.
// The ref identity is used to determine when sensitive fields
// have changed and need to be (re)encrypted.
func (r *Identity) Encrypt(ref *Identity) (err error) {
	passphrase := Settings.Encryption.Passphrase
	aes := encryption.New(passphrase)
	if r.Password != ref.Password {
		if r.Password != "" {
			r.Password, err = aes.Encrypt(r.Password)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	if r.Key != ref.Key {
		if r.Key != "" {
			r.Key, err = aes.Encrypt(r.Key)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	if r.Settings != ref.Settings {
		if r.Settings != "" {
			r.Settings, err = aes.Encrypt(r.Settings)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}

//
// Decrypt sensitive fields.
func (r *Identity) Decrypt() (err error) {
	passphrase := Settings.Encryption.Passphrase
	aes := encryption.New(passphrase)
	if r.Password != "" {
		r.Password, err = aes.Decrypt(r.Password)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	if r.Key != "" {
		r.Key, err = aes.Decrypt(r.Key)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	if r.Settings != "" {
		r.Settings, err = aes.Decrypt(r.Settings)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}
