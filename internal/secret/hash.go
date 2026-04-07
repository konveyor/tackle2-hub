package secret

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	hmacPrefix = "hmac-sha256:"
)

// Hash hashes an API key secret using HMAC-SHA256.
func Hash(secret string) (hashed string) {
	if len(secret) == 0 || isHashed(secret) {
		hashed = secret
		return
	}
	h := hmac.New(sha256.New, []byte(Settings.Passphrase))
	h.Write([]byte(secret))
	hash := h.Sum(nil)
	hashed = base64.URLEncoding.EncodeToString(hash)
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

// HashPassword hashes a password using bcrypt and prevents double hashing.
// Returns the bcrypt hash of the password, or the input unchanged if already hashed.
func HashPassword(password string) (hashed string, err error) {
	if len(password) == 0 {
		hashed = password
		return
	}
	p := []byte(password)
	_, err = bcrypt.Cost(p)
	if err == nil {
		hashed = password
		return
	}
	hash, err := bcrypt.GenerateFromPassword(p, bcrypt.DefaultCost)
	if err != nil {
		return
	}
	hashed = string(hash)
	return
}

// MatchPassword compares a plaintext password against a bcrypt hash.
// Returns true if the password matches the hash.
func MatchPassword(password string, hashed string) (matched bool, err error) {
	p := []byte(password)
	h := []byte(hashed)
	err = bcrypt.CompareHashAndPassword(h, p)
	if err == nil {
		matched = true
		return
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		err = nil
		return
	}
	return
}
