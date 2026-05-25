package secret

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var Settings = &settings.Settings.Encryption

// Cipher implements encryption.
type Cipher interface {
	Use(string)
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

// Encrypt object.
// When object is:
// - *string - the string is encrypted.
// - struct - (string) fields with `secret:` tag are encrypted.
// - map[string]any - string fields are encrypted.
func Encrypt(object any) (err error) {
	cipher := &AESGCM{}
	cipher.Use(Settings.Passphrase)
	secret := Secret{Cipher: cipher}
	err = secret.Encrypt(object)
	err = liberr.Wrap(err)
	return
}

// Decrypt object.
// When object is:
// - *string - the string is decrypted.
// - struct - (string) fields with `secret:` tag are decrypted.
// - map[string]any - string fields are decrypted.
func Decrypt(object any) (err error) {
	cipher := &AESGCM{}
	cipher.Use(Settings.Passphrase)
	secret := Secret{Cipher: cipher}
	err = secret.Decrypt(object)
	err = liberr.Wrap(err)
	return
}

// Encode object.
// When object is:
// - *string - the string is encrypted.
// - struct - (string) fields with `secret:` tag are encoded based on tag (value).
// - map[string]any - string fields are encrypted.
func Encode(object any) (fields []Field, err error) {
	cipher := &AESGCM{}
	cipher.Use(Settings.Passphrase)
	secret := Secret{Cipher: cipher}
	fields, err = secret.Encode(object)
	err = liberr.Wrap(err)
	return
}

// Decode object.
// When object is:
// - *string - the string is decrypted.
// - struct - (string) fields with `secret:` tag are decoded based on tag (value).
// - map[string]any - string fields are decrypted.
func Decode(object any) (err error) {
	cipher := &AESGCM{}
	cipher.Use(Settings.Passphrase)
	secret := Secret{Cipher: cipher}
	err = secret.Decode(object)
	err = liberr.Wrap(err)
	return
}
