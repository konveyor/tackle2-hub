package auth

import (
	"sync"
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewCache returns a configured cache.
func NewCache(db *gorm.DB) (cache *KeyCache) {
	cache = &KeyCache{db: db}
	cache.reset()
	return
}

// KeyCache provides an APIKey cache.
// Keys are cached to mitigate DB pressure during heavy loads.
type KeyCache struct {
	db        *gorm.DB
	mutex     sync.RWMutex
	bySecret  map[string]APIKey
	byDigest  map[string]APIKey
	resetLast time.Time
}

func (r *KeyCache) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reset()
}

// Delete a key by digest.
func (r *KeyCache) Delete(digest string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	key, found := r.byDigest[digest]
	if found {
		delete(r.bySecret, key.Secret)
		delete(r.byDigest, digest)
	}
}

func (r *KeyCache) Get(secret string) (key APIKey, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if time.Since(r.resetLast) >
		Settings.Auth.APIKey.CacheLifespan {
		r.reset()
	}
	key, found := r.bySecret[secret]
	if found {
		return
	}
	digest := hashSecret(secret)
	m := &model.APIKey{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("User.Roles")
	db = db.Preload("User.Roles.Permissions")
	db = db.Where("digest", digest)
	db = db.Where("expiration > ?", time.Now())
	err = db.First(m).Error
	if err != nil {
		err = &NotAuthenticated{
			Token: secret,
		}
		return
	}
	key.Secret = secret
	key.Digest = digest
	key.Expiration = m.Expiration
	if m.UserID != nil {
		if m.User == nil {
			err = &NotAuthenticated{
				Token: secret,
			}
			return
		}
		key.User = m.User.UUID
		unique := make(map[string]byte)
		for _, u := range m.User.Roles {
			for _, perm := range u.Permissions {
				unique[perm.Scope] = 0
			}
		}
		for scope, _ := range unique {
			key.Scopes = append(key.Scopes, scope)
		}
	}
	if m.TaskID != nil {
		if m.Task == nil {
			err = &NotAuthenticated{
				Token: secret,
			}
			return
		}
		switch m.Task.State {
		case task.Succeeded,
			task.Failed,
			task.Canceled:
			err = &NotAuthenticated{
				Token: secret,
			}
			return
		}
	}
	r.bySecret[key.Secret] = key
	r.byDigest[key.Digest] = key
	return
}

func (r *KeyCache) reset() {
	r.bySecret = make(map[string]APIKey)
	r.byDigest = make(map[string]APIKey)
	r.resetLast = time.Now()
}
