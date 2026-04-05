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

type KeyCache struct {
	db        *gorm.DB
	mutex     sync.RWMutex
	content   map[string]APIKey
	resetLast time.Time
}

func (r *KeyCache) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reset()
}

func (r *KeyCache) Get(nakedSecret string) (key APIKey, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if time.Since(r.resetLast) >
		Settings.Auth.APIKey.CacheLifespan {
		r.reset()
	}
	key, found := r.content[nakedSecret]
	if found {
		return
	}
	m := &model.APIKey{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("User.Roles")
	db = db.Preload("User.Roles.Permissions")
	db = db.Where("digest", hashSecret(nakedSecret))
	db = db.Where("expiration > ?", time.Now())
	err = db.First(m).Error
	if err != nil {
		err = &NotAuthenticated{
			Token: nakedSecret,
		}
		return
	}
	key.Secret = nakedSecret
	key.Expiration = m.Expiration
	if m.UserID != nil {
		if m.User == nil {
			err = &NotAuthenticated{
				Token: nakedSecret,
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
				Token: nakedSecret,
			}
			return
		}
		switch m.Task.State {
		case task.Succeeded,
			task.Failed,
			task.Canceled:
			err = &NotAuthenticated{
				Token: nakedSecret,
			}
			return
		}
	}
	r.content[nakedSecret] = key
	return
}

func (r *KeyCache) reset() {
	r.content = make(map[string]APIKey)
	r.resetLast = time.Now()
}
