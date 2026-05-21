package secret

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// Hash hashes a password using hmac-sha256 and wraps it in URL-safe base64.
// Returns the encoded hmac-sha256 digest, or the input unchanged if already hashed.
func Hash(secret string) (hashed string) {
	if len(secret) == 0 || isHashed(secret) {
		hashed = secret
		return
	}
	h := hmac.New(sha256.New, []byte(Settings.Passphrase))
	h.Write([]byte(secret))
	digest := h.Sum(nil)
	hashed = base64.URLEncoding.EncodeToString(digest)
	return
}

// isHashed s true when already hashed.
func isHashed(s string) (hashed bool) {
	if len(s) != 44 {
		return
	}
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return
	}
	hashed = len(decoded) == 32
	return
}

// isHashedPassword returns true when already in bcrypt hashed format.
func isHashedPassword(s string) (hashed bool) {
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return
	}
	_, err = bcrypt.Cost(decoded)
	hashed = err == nil
	return
}

// HashPassword hashes a password using bcrypt and wraps it in URL-safe base64.
// Returns the encoded bcrypt digest, or the input unchanged if already hashed.
// Passwords longer than 72 bytes are truncated to 72 bytes due to bcrypt limitations.
func HashPassword(password string) (hashed string) {
	if len(password) == 0 {
		hashed = password
		return
	}
	if isHashedPassword(password) {
		hashed = password
		return
	}
	p := []byte(password)
	if len(p) > 72 {
		p = p[:72]
	}
	digest, err := bcrypt.GenerateFromPassword(p, bcrypt.DefaultCost)
	if err != nil {
		return
	}
	hashed = base64.URLEncoding.EncodeToString(digest)
	return
}

// MatchPassword compares a plaintext password against an encoded bcrypt hash.
// Returns true if the password matches the hash.
func MatchPassword(password string, hashed string) (matched bool) {
	digest, err := base64.URLEncoding.DecodeString(hashed)
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
