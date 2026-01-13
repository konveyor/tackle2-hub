package secret

import (
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
	return
}
