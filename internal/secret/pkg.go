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
	_, err = secret.Decode(object)
	err = liberr.Wrap(err)
	return
}

// Fields returns secret fields.
func Fields(object any) (fields []Field, err error) {
	fields, err = Secret{}.Fields(object)
	return
}

// Redact updates the value of secret fields with a `mask`.
func Redact(object any, mask string) (err error) {
	err = Secret{}.Redact(object, mask)
	return
}

// RevertRedacted reverts redacted fields (defined by mask).
// When a secret field Secret() matches the mask, it is updated with the
// corresponding field in (other) fields.
func RevertRedacted(fields []Field, other any, mask string) (err error) {
	err = Secret{}.RevertRedacted(fields, other, mask)
	return
}
