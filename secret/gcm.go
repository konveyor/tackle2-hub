package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// AESGCM Galois/Counter Mode encryption.
type AESGCM struct {
	// Key Length must be (16|24|32).
	Key []byte
}

// Use generate the key based on the passphrase.
// The key is padded to be 32 bits.
// The padded by repeating the passphrase.
func (r *AESGCM) Use(passphrase string) {
	passphrase = strings.TrimSpace(passphrase)
	pLen := len(passphrase)
	if pLen == 0 {
		return
	}
	i := 0
	r.Key = []byte{}
	for {
		if i == pLen {
			i = 0
		}
		r.Key = append(r.Key, passphrase[i])
		if len(r.Key) == 32 {
			break
		}
		i++
	}
}

// Encrypt plain string.
// Returns an AESCFB-GCM encrypted; base64 encoded string.
func (r *AESGCM) Encrypt(plain string) (encrypted string, err error) {
	if plain == "" {
		encrypted = plain
		return
	}
	block, err := aes.NewCipher(r.Key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return
	}
	b := gcm.Seal(nonce, nonce, []byte(plain), nil)
	encrypted = r.encode(b)
	return
}

// Decrypt and AESCFB encrypted string.
// The `encrypted` string is an AESCFB-GCM encrypted; base64 encoded string.
// Returns the decoded string.
func (r *AESGCM) Decrypt(encrypted string) (plain string, err error) {
	if encrypted == "" {
		plain = encrypted
		return
	}
	block, err := aes.NewCipher(r.Key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	b, err := r.decode(encrypted)
	if err != nil {
		return
	}
	nonceSize := gcm.NonceSize()
	if len(b) < nonceSize {
		err = fmt.Errorf("input not valid")
		return
	}
	nonce := b[:nonceSize]
	b = b[nonceSize:]
	b, err = gcm.Open(nil, nonce, b, nil)
	if err != nil {
		return
	}
	plain = string(b)
	return
}

// encode string.
func (r *AESGCM) encode(in []byte) (out string) {
	out = base64.StdEncoding.EncodeToString(in)
	return
}

// decode bytes.
func (r *AESGCM) decode(in string) (out []byte, err error) {
	out, err = base64.StdEncoding.DecodeString(in)
	return
}
