package auth

import (
	"strings"
	"sync"
	"time"

	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewCache returns a cache.
func NewCache(db *gorm.DB) (cache *TokenCache) {
	cache = &TokenCache{db: db}
	cache.reset()
	return
}

// TokenCache provides an Token cache.
// Tokens are cached to mitigate DB pressure during heavy loads.
type TokenCache struct {
	db        *gorm.DB
	mutex     sync.RWMutex
	byId      map[uint]Token
	byDigest  map[string]Token
	resetLast time.Time
}

func (r *TokenCache) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reset()
}

// Delete a token by id.
func (r *TokenCache) Delete(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	token, found := r.byId[id]
	if found {
		delete(r.byDigest, token.Digest)
		delete(r.byId, id)
	}
}

// Get returns a token by secret.
func (r *TokenCache) Get(tokenSecret string) (m Token, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if time.Since(r.resetLast) >
		Settings.Auth.APIKey.CacheLifespan {
		r.reset()
	}
	digest := secret.Hash(tokenSecret)
	m, found := r.byDigest[digest]
	if found {
		return
	}
	m = Token{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("User.Roles")
	db = db.Preload("User.Roles.Permissions")
	db = db.Where("digest", digest)
	db = db.Where("expiration > ?", time.Now())

	err = db.First(&m).Error
	if err != nil {
		err = &NotAuthenticated{}
		return
	}
	if m.UserID != nil {
		if m.User == nil {
			err = &NotAuthenticated{}
			return
		}
		m.Subject = m.User.Subject
		unique := make(map[string]bool)
		scopes := make([]string, 0)
		for _, u := range m.User.Roles {
			for _, perm := range u.Permissions {
				if !unique[perm.Scope] {
					scopes = append(scopes, perm.Scope)
				}
				unique[perm.Scope] = true
			}
		}
		m.Scopes = strings.Join(scopes, " ")
	}
	if m.TaskID != nil {
		if m.Task == nil {
			err = &NotAuthenticated{}
			return
		}
		switch m.Task.State {
		case task.Succeeded,
			task.Failed,
			task.Canceled:
			err = &NotAuthenticated{}
			return
		default:
			strings.Join(AddonScopes, " ")
		}
	}
	r.byId[m.ID] = m
	r.byDigest[digest] = m
	return
}

func (r *TokenCache) reset() {
	r.byId = make(map[uint]Token)
	r.byDigest = make(map[string]Token)
	r.resetLast = time.Now()
}
