package secret

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	hmacPrefix   = "hmac-sha256:"
	bcryptPrefix = "bcrypt:"
)

// Hash hashes a password using hmac-sha256 and wraps it in URL-safe base64.
// Returns the encoded hmac-sha256 hash with prefix, or the input unchanged if already hashed.
func Hash(secret string) (hashed string) {
	if len(secret) == 0 || isHashed(secret) {
		hashed = secret
		return
	}
	h := hmac.New(sha256.New, []byte(Settings.Passphrase))
	h.Write([]byte(secret))
	digest := h.Sum(nil)
	hashed = base64.URLEncoding.EncodeToString(digest)
	hashed = hmacPrefix + hashed
	return
}

// isHashed returns true when already hashed.
func isHashed(s string) (hashed bool) {
	if !strings.HasPrefix(s, hmacPrefix) {
		return
	}
	encoded := strings.TrimPrefix(s, hmacPrefix)
	if len(encoded) != 44 {
		return
	}
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return
	}
	hashed = len(decoded) == 32
	return
}

// isHashedPassword returns true when already in bcrypt hashed format.
func isHashedPassword(s string) (hashed bool) {
	if !strings.HasPrefix(s, bcryptPrefix) {
		return
	}
	encoded := strings.TrimPrefix(s, bcryptPrefix)
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return
	}
	_, err = bcrypt.Cost(decoded)
	hashed = (err == nil)
	return
}

// HashPassword hashes a password using bcrypt and wraps it in URL-safe base64.
// Returns the encoded bcrypt hash with prefix, or the input unchanged if already hashed.
func HashPassword(password string) (hashed string, err error) {
	if len(password) == 0 {
		hashed = password
		return
	}
	if isHashedPassword(password) {
		hashed = password
		return
	}
	p := []byte(password)
	digest, err := bcrypt.GenerateFromPassword(p, bcrypt.DefaultCost)
	if err != nil {
		return
	}
	encoded := base64.URLEncoding.EncodeToString(digest)
	hashed = bcryptPrefix + encoded
	return
}

// MatchPassword compares a plaintext password against an encoded bcrypt hash.
// Returns true if the password matches the hash.
func MatchPassword(password string, hashed string) (matched bool) {
	if !strings.HasPrefix(hashed, bcryptPrefix) {
		return
	}
	encoded := strings.TrimPrefix(hashed, bcryptPrefix)
	digest, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return
	}
	p := []byte(password)
	err = bcrypt.CompareHashAndPassword(digest, p)
	if err == nil {
		matched = true
		return
	}
	return
}
