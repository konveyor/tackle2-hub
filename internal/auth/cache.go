package auth

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
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
	db            *gorm.DB
	mutex         sync.RWMutex
	permById      map[uint]*Permission
	roleById      map[uint]*Role
	userById      map[uint]*User
	taskById      map[uint]*Task
	identById     map[uint]*Identity
	tokenById     map[uint]*Token
	tokenByDigest map[string]*Token
	refreshed     time.Time
}

func (r *Cache) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reset()
}

func (r *Cache) Refresh() (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	err = r.refresh()
	return
}

func (r *Cache) RoleSaved(m *Role) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.roleById[m.ID] = m
}

func (r *Cache) RoleDeleted(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.roleById, id)
}

func (r *Cache) UserSaved(m *User) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.userById[m.ID] = m
}

func (r *Cache) UserDeleted(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.userById, id)
}

func (r *Cache) TaskSaved(m *Task) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.taskById[m.ID] = m
}

func (r *Cache) TaskDeleted(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.taskById, id)
}

func (r *Cache) IdentitySaved(m *Identity) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.identById[m.ID] = m
}

func (r *Cache) IdentityDeleted(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.identById, id)
}

func (r *Cache) TokenSaved(m *Token) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.tokenById[m.ID] = m
	r.tokenByDigest[m.Digest] = m
}

func (r *Cache) TokenDeleted(id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	token, found := r.tokenById[id]
	if found {
		delete(r.tokenByDigest, token.Digest)
		delete(r.tokenById, id)
	}
}

// GetToken returns a PAT.
func (r *Cache) GetToken(token string) (m *Token, err error) {
	var needsRefresh bool
	func() {
		r.mutex.RLock()
		defer r.mutex.RUnlock()
		needsRefresh = time.Since(r.refreshed) > Settings.CacheLifespan
		if !needsRefresh {
			m, err = r.getToken(token)
			if errors.Is(err, &NotFound{}) {
				needsRefresh = true
			}
		}
	}()
	if needsRefresh {
		func() {
			r.mutex.Lock()
			defer r.mutex.Unlock()
			err = r.refresh()
		}()
		func() {
			r.mutex.RLock()
			defer r.mutex.RUnlock()
			if err == nil {
				m, err = r.getToken(token)
			}
		}()
	}
	return
}

func (r *Cache) reset() {
	r.permById = make(map[uint]*Permission)
	r.roleById = make(map[uint]*Role)
	r.userById = make(map[uint]*User)
	r.taskById = make(map[uint]*Task)
	r.tokenById = make(map[uint]*Token)
	r.identById = make(map[uint]*Identity)
	r.tokenByDigest = make(map[string]*Token)
}

func (r *Cache) refresh() (err error) {
	r.reset()
	err = r.getPerms()
	if err != nil {
		return
	}
	err = r.getRoles()
	if err != nil {
		return
	}
	err = r.getUsers()
	if err != nil {
		return
	}
	err = r.getTasks()
	if err != nil {
		return
	}
	err = r.getIdentities()
	if err != nil {
		return
	}
	err = r.getTokens()
	if err != nil {
		return
	}
	r.refreshed = time.Now()
	return
}

func (r *Cache) getPerms() (err error) {
	list := make([]*Permission, 0)
	err = r.db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.permById[m.ID] = m
	}
	return
}

func (r *Cache) getRoles() (err error) {
	list := make([]*Role, 0)
	db := r.db.Preload(clause.Associations)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.roleById[m.ID] = m
	}
	return
}

func (r *Cache) getUsers() (err error) {
	list := make([]*User, 0)
	db := r.db.Preload(clause.Associations)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.userById[m.ID] = m
	}
	return
}

func (r *Cache) getTasks() (err error) {
	list := make([]*Task, 0)
	err = r.db.Find(
		&list,
		"state IN ?",
		[]string{
			task.Pending,
			task.Running,
		}).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.taskById[m.ID] = m
	}
	return
}

func (r *Cache) getIdentities() (err error) {
	list := make([]*Identity, 0)
	err = r.db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.identById[m.ID] = m
	}
	return
}

func (r *Cache) getTokens() (err error) {
	list := []*Token{}
	db := r.db.Preload(clause.Associations)
	db = db.Where("kind", KindAPIKey)
	db = db.Where("expiration > ?", time.Now())
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.tokenById[m.ID] = m
		r.tokenByDigest[m.Digest] = m
	}
	return
}

// getToken returns a PAT.
func (r *Cache) getToken(token string) (m *Token, err error) {
	cached, found := r.tokenByDigest[secret.Hash(token)]
	if !found {
		err = &NotFound{
			Resource: "token",
			Id:       token,
		}
		return
	}
	// Create a copy to avoid modifying cached instance
	m = &Token{Token: cached.Token}
	// user binding.
	if m.UserID != nil {
		user, found := r.userById[*m.UserID]
		if !found {
			err = &NotFound{
				Resource: "user",
				Id:       strconv.Itoa(int(*m.UserID)),
			}
			return
		}
		m.Subject = user.Subject
		m.Scopes = strings.Join(user.scopes(r.roleById), " ")
		return
	}
	// task binding.
	if m.TaskID != nil {
		task, found := r.taskById[*m.TaskID]
		if !found {
			err = &NotFound{
				Resource: "task",
				Id:       strconv.Itoa(int(*m.TaskID)),
			}
			return
		}
		m.Subject = "task:" + strconv.Itoa(int(task.ID))
		m.Scopes = strings.Join(AddonScopes, " ")
		return
	}
	// IdP identity binding.
	if m.IdpIdentityID != nil {
		identity, found := r.identById[*m.IdpIdentityID]
		if !found {
			err = &NotFound{
				Resource: "identity",
				Id:       strconv.Itoa(int(*m.IdpIdentityID)),
			}
			return
		}
		m.Subject = identity.Subject
		m.Scopes = identity.Scopes
		return
	}
	return
}

// User alias.
type User model.User

// scopes returns the user's scopes.
func (m *User) scopes(roles map[uint]*Role) (scopes []string) {
	scopeMap := make(map[string]bool)
	for _, r := range m.Roles {
		role, found := roles[r.ID]
		if !found {
			continue
		}
		for _, scope := range role.scopes() {
			if !scopeMap[scope] {
				scopes = append(scopes, scope)
				scopeMap[scope] = true
			}
		}
	}
	sort.Strings(scopes)
	return
}

// Role alias.
type Role model.Role

// scopes returns the roles scopes.
func (m *Role) scopes() (scopes []string) {
	scopeMap := make(map[string]bool)
	for _, p := range m.Permissions {
		if !scopeMap[p.Scope] {
			scopes = append(scopes, p.Scope)
			scopeMap[p.Scope] = true
		}
	}
	sort.Strings(scopes)
	return
}

// Permission alias.
type Permission = model.Permission

// Task alias.
type Task = model.Task

// Identity alias.
type Identity = model.IdpIdentity

// Grant alias.
type Grant = model.Grant

// Token alias.
type Token struct {
	model.Token
	Secret string `gorm:"-"`
}
