package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// AESCFB Cipher Feedback (stream-like) encryption.
// Deprecated
type AESCFB struct {
	// Key Length must be (12|24|32).
	Key []byte
}

// Use Sets the key using the passphrase.
// Only the first 32 bytes of the passphrase are used.
func (r *AESCFB) Use(passphrase string) {
	keyLen := 32
	r.Key = make([]byte, keyLen)
	input := []byte(passphrase)
	for n := range input {
		if n < keyLen {
			r.Key[n] = input[n]
		} else {
			break
		}
	}
}

// Encrypt plain string.
// Returns an AESCFB encrypted; base64 encoded string.
func (r *AESCFB) Encrypt(plain string) (encrypted string, err error) {
	if plain == "" {
		encrypted = plain
		return
	}
	block, err := aes.NewCipher(r.Key)
	if err != nil {
		return
	}
	input := make([]byte, aes.BlockSize+len(plain))
	iv := input[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(input[aes.BlockSize:], []byte(plain))
	encrypted = r.encode(input)
	return
}

// Decrypt and AESCFB encrypted string.
// The `encrypted` string is an AESCFB encrypted; base64 encoded string.
// Returns the decoded string.
func (r *AESCFB) Decrypt(encrypted string) (plain string, err error) {
	if encrypted == "" {
		plain = encrypted
		return
	}
	block, err := aes.NewCipher(r.Key)
	if err != nil {
		return
	}
	input, err := r.decode(encrypted)
	if err != nil {
		return
	}
	if len(input) < aes.BlockSize {
		return
	}
	iv := input[:aes.BlockSize]
	input = input[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(input, input)
	plain = string(input)
	return
}

// encode string.
func (r *AESCFB) encode(in []byte) (out string) {
	out = base64.StdEncoding.EncodeToString(in)
	return
}

// decode bytes.
func (r *AESCFB) decode(in string) (out []byte, err error) {
	out, err = base64.StdEncoding.DecodeString(in)
	return
}
