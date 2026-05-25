package cache

import (
	"strconv"
	"strings"
	"sync"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// New returns a cache.
func New(db *gorm.DB) (cache *Cache) {
	cache = &Cache{db: db}
	cache.reset()
	return
}

// Cache caches resources.
// Tokens are cached to mitigate DB pressure during heavy loads.
//
// Cache Strategy:
//   - Notifications: Saved/Deleted methods immediately update the cache
//   - Safety-net: Periodic refresh.
//   - Password changes, role updates, and other changes are propagated
//     immediately via notifications
type Cache struct {
	db              *gorm.DB
	mutex           sync.RWMutex
	permById        map[uint]*Permission
	roleById        map[uint]*Role
	roleByName      map[string]*Role
	userById        map[uint]*User
	userBySubject   map[string]*User
	userByLogin     map[string]*User
	taskById        map[uint]*Task
	identById       map[uint]*Identity
	identBySubject  map[string]*Identity
	identByLogin    map[string]*Identity
	clientById      map[uint]*IdpClient
	clientBySubject map[string]*IdpClient
	tokenById       map[uint]*Token
	tokenByDigest   map[string]*Token
	refreshed       time.Time
}

// Reset clears all cached data.
func (r *Cache) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.reset()
}

// Refresh reloads all data from the database.
func (r *Cache) Refresh() (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	err = r.refresh()
	return
}

// RoleSaved updates the cache when a role is saved.
func (r *Cache) RoleSaved(m *Role) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.RoleSaved(m)
		return
	})
}

// RoleDeleted removes a role from the cache.
func (r *Cache) RoleDeleted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.RoleDeleted(id)
		return
	})
}

// UserSaved updates the cache when a user is saved.
func (r *Cache) UserSaved(m *User) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.UserSaved(m)
		return
	})
}

// UserDeleted removes a user from the cache.
func (r *Cache) UserDeleted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.UserDeleted(id)
		return
	})
}

// TaskGranted task token granted.
func (r *Cache) TaskGranted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.TaskGranted(id)
		return
	})
}

// TaskRevoked removes a task from the cache.
func (r *Cache) TaskRevoked(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.TaskRevoked(id)
		return
	})
}

// IdentitySaved updates the cache when an identity is saved.
func (r *Cache) IdentitySaved(m *Identity) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.IdentitySaved(m)
		return
	})
}

// IdentityDeleted removes an identity from the cache.
func (r *Cache) IdentityDeleted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.IdentityDeleted(id)
		return
	})
}

// ClientSaved updates the cache when a client is saved.
func (r *Cache) ClientSaved(m *IdpClient) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.ClientSaved(m)
		return
	})
}

// ClientDeleted removes a client from the cache.
func (r *Cache) ClientDeleted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.ClientDeleted(id)
		return
	})
}

// TokenSaved updates the cache when a token is saved.
func (r *Cache) TokenSaved(m *Token) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.TokenSaved(m)
		return
	})
}

// TokenDeleted removes a token from the cache.
func (r *Cache) TokenDeleted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.TokenDeleted(id)
		return
	})
}

// FindToken returns a PAT.
func (r *Cache) FindToken(token string) (m *Token, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	m, err = r.getToken(token)
	return
}

// FindSubject returns a subject.
func (r *Cache) FindSubject(subject string) (m *Subject, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	var found bool
	m, found = r.findSubject(subject)
	if !found {
		err = &NotFound{
			Resource: "subject",
			Id:       subject,
		}
	}
	return
}

// FindIdentityByLogin finds and returns identities by login.
func (r *Cache) FindIdentityByLogin(login string) (m *Identity, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	m, found := r.identByLogin[login]
	if !found {
		err = &NotFound{
			Resource: "identity",
			Id:       login,
		}
	}
	return
}

// FindRoleById return a role by id.
func (r *Cache) FindRoleById(id uint) (m *Role, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	m, found := r.roleById[id]
	if !found {
		err = &NotFound{
			Resource: "Role",
			Id:       strconv.Itoa(int(id)),
		}
	}
	return
}

// FindRoleByName return a role by name.
func (r *Cache) FindRoleByName(name string) (m *Role, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	m, found := r.roleByName[name]
	if !found {
		err = &NotFound{
			Resource: "Role",
			Id:       name,
		}
	}
	return
}

// FindUserByLogin returns a user by login.
func (r *Cache) FindUserByLogin(login string) (m *User, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	m, found := r.userByLogin[login]
	if !found {
		err = &NotFound{
			Resource: "user",
			Id:       login,
		}
	}
	return
}

// FindTaskById returns a task by ID.
func (r *Cache) FindTaskById(id uint) (m *Task, err error) {
	err = r.ensureFresh()
	if err != nil {
		return
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	m, found := r.taskById[id]
	if !found {
		err = &NotFound{
			Resource: "task",
			Id:       strconv.Itoa(int(id)),
		}
	}
	return
}

// ensureFresh refreshes the cache if CacheLifespan has elapsed.
func (r *Cache) ensureFresh() (err error) {
	var needsRefresh bool
	func() {
		r.mutex.RLock()
		defer r.mutex.RUnlock()
		needsRefresh = time.Since(r.refreshed) > Settings.CacheLifespan
	}()
	if needsRefresh {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		if time.Since(r.refreshed) > Settings.CacheLifespan {
			err = r.refresh()
		}
	}
	return
}

// Begin returns a cache transaction.
// The transaction holds a changelog of cache operations.
// Changes are applied atomically on Commit().
func (r *Cache) Begin() (tx *Tx) {
	tx = &Tx{
		cache:   r,
		changes: make([]func(), 0),
	}
	return
}

// Transaction executes a function within a cache transaction.
// Commits on success, rolls back on error.
func (r *Cache) Transaction(fn func(*Tx) error) (err error) {
	tx := r.Begin()
	defer tx.Rollback()
	err = fn(tx)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

// reset allocate cache content backing maps.
func (r *Cache) reset() {
	r.permById = make(map[uint]*Permission)
	r.roleById = make(map[uint]*Role)
	r.roleByName = make(map[string]*Role)
	r.userById = make(map[uint]*User)
	r.userBySubject = make(map[string]*User)
	r.userByLogin = make(map[string]*User)
	r.taskById = make(map[uint]*Task)
	r.tokenById = make(map[uint]*Token)
	r.identById = make(map[uint]*Identity)
	r.identBySubject = make(map[string]*Identity)
	r.identByLogin = make(map[string]*Identity)
	r.clientById = make(map[uint]*IdpClient)
	r.clientBySubject = make(map[string]*IdpClient)
	r.tokenByDigest = make(map[string]*Token)
	r.refreshed = time.Time{}
}

// refresh cached content.
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
	err = r.getClients()
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

// getPerms fetches permissions from the DB and populates.
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

// getRoles fetches roles from the DB and populates.
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
		r.roleByName[m.Name] = m
	}
	return
}

// getUsers fetches users from the DB and populates.
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
		r.userBySubject[m.Subject] = m
		r.userByLogin[m.Login] = m
	}
	return
}

// getTasks fetches tasks state=(pending|running) from the DB and populates.
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

// getIdentities fetches idp identities from the DB and populates.
func (r *Cache) getIdentities() (err error) {
	list := make([]*Identity, 0)
	err = r.db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		err = secret.Decrypt(m)
		if err == nil {
			r.identById[m.ID] = m
			r.identBySubject[m.Subject] = m
			r.identByLogin[m.Login] = m
		} else {
			return
		}
	}
	return
}

// getClients fetches idp clients from the DB and populates.
func (r *Cache) getClients() (err error) {
	list := make([]*IdpClient, 0)
	err = r.db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		r.clientById[m.ID] = m
		r.clientBySubject[m.Subject] = m
	}
	return
}

// getTokens fetches permissions from the DB and populates.
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
		m.Scopes = strings.Join(user.GetScopes(r), " ")
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
	// IdP client binding.
	if m.IdpClientID != nil {
		client, found := r.clientById[*m.IdpClientID]
		if !found {
			err = &NotFound{
				Resource: "client",
				Id:       strconv.Itoa(int(*m.IdpClientID)),
			}
			return
		}
		m.Subject = client.Subject
		m.Scopes = strings.Join(client.GetScopes(), " ")
		return
	}
	return
}

// findSubject returns the subject.
func (r *Cache) findSubject(subject string) (s *Subject, found bool) {
	user, found := r.userBySubject[subject]
	if found {
		s = &Subject{}
		s.WithUser(user, r)
		return
	}
	identity, found := r.identBySubject[subject]
	if found {
		s = &Subject{}
		s.WithIdentity(identity)
		return
	}
	client, found := r.clientBySubject[subject]
	if found {
		s = &Subject{}
		s.WithClient(client, r)
		return
	}
	return
}
