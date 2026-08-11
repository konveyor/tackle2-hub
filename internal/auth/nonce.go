package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// NonceCache provides nonce generation and authentication.
type NonceCache struct {
	mutex  sync.Mutex
	issued map[string]time.Time
}

// Issue returns a generated nonce.
func (n *NonceCache) Issue(d time.Duration) (nonce string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.prune()
	nonce = n.nonce()
	expiry := time.Now().Add(d)
	n.issued[nonce] = expiry
	return
}

// Redeem validates the nonce and deletes it.
func (n *NonceCache) Redeem(nonce string) (err error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.prune()
	if !Settings.Auth.Required {
		return
	}
	expiry, found := n.issued[nonce]
	if found {
		delete(n.issued, nonce)
		if time.Now().Before(expiry) {
			return
		}
	}
	err = &NotAuthenticated{
		Reason: "nonce not valid.",
		Token:  nonce,
	}
	return
}

// nonce returns a new generated nonce.
func (n *NonceCache) nonce() (s string) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	s = hex.EncodeToString(b)
	return
}

// prune delete expired nonces.
func (n *NonceCache) prune() {
	if n.issued == nil {
		n.issued = make(map[string]time.Time)
	}
	for nonce, expiry := range n.issued {
		if time.Now().After(expiry) {
			delete(n.issued, nonce)
		}
	}
}
