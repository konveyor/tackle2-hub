package encryption

import "github.com/konveyor/tackle2-hub/settings"

var Settings = &settings.Settings.Encryption

type Cipher interface {
	Use(string)
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

func Encrypt(object any) (err error) {
	cipher := &AESGCM{}
	cipher.Use(Settings.Passphrase)
	secret := Secret{Cipher: cipher}
	err = secret.Encrypt(object)
	return
}

func Decrypt(object any) (err error) {
	cipher := &AESGCM{}
	cipher.Use(Settings.Passphrase)
	secret := Secret{Cipher: cipher}
	err = secret.Decrypt(object)
	return
}
