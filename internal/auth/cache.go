package auth

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewCache returns a cache.
func NewCache(db *gorm.DB) (cache *Cache) {
	cache = &Cache{db: db}
	cache.reset()
	return
}

// Cache caches resources.
// Tokens are cached to mitigate DB pressure during heavy loads.
type Cache struct {
	db        *gorm.DB
	mutex     sync.RWMutex
	byId      map[uint]*Token
	byDigest  map[string]*Token
	resetLast time.Time
}

func (r *Cache) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reset()
}

// Delete a token by id.
func (r *Cache) Delete(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	token, found := r.byId[id]
	if found {
		delete(r.byDigest, token.Digest)
		delete(r.byId, id)
	}
}

// GetPAT returns a PAT.
func (r *Cache) GetPAT(token string) (m *Token, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if time.Since(r.resetLast) >
		Settings.Auth.APIKey.CacheLifespan {
		r.reset()
	}
	digest := secret.Hash(token)
	m, found := r.byDigest[digest]
	if found {
		return
	}
	m = &Token{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("User.Roles")
	db = db.Preload("User.Roles.Permissions")
	db = db.Where("digest", digest)
	db = db.Where("expiration > ?", time.Now())
	db = db.Where("kind", KindAPIKey)
	err = db.First(&m).Error
	if err != nil {
		err = &NotFound{
			Resource: "PAT",
			Id:       token,
		}
		return
	} else {
		found = true
	}
	err = r.putPAT(digest, m)
	return
}

// putPAT adds the token to the cache with inherited scopes.
func (r *Cache) putPAT(digest string, m *Token) (err error) {
	defer func() {
		if err == nil {
			r.byId[m.ID] = m
			r.byDigest[digest] = m
		}
	}()
	//
	// PAT owned by a user.
	if m.UserID != nil {
		id := strconv.Itoa(int(*m.UserID))
		if m.User == nil {
			err = &NotFound{
				Resource: "User",
				Id:       id,
			}
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
		return
	}
	// API-Key owned by task.
	if m.TaskID != nil {
		id := strconv.Itoa(int(*m.TaskID))
		if m.Task == nil {
			err = &NotFound{
				Resource: "Task",
				Id:       id,
			}
			return
		}
		switch m.Task.State {
		case task.Succeeded,
			task.Failed,
			task.Canceled:
			err = &NotFound{
				Resource: "Task",
				Id:       id,
			}
			return
		default:
			m.Scopes = strings.Join(AddonScopes, " ")
			m.Subject = "task:" + id
		}
		return
	}
	//
	// PAT owned by a (remote) IdP identity.
	if m.IdpIdentity != nil {
		m.Subject = m.IdpIdentity.Subject
		m.Scopes = m.IdpIdentity.Scopes
	}
	return
}

func (r *Cache) reset() {
	r.byId = make(map[uint]*Token)
	r.byDigest = make(map[string]*Token)
	r.resetLast = time.Now()
}
