package secret

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
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
