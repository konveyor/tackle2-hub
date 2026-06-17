package cache

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	Log = logr.WithName("auth.cache")
)

// New returns a cache.
func New(db *gorm.DB) (cache *Cache) {
	d := Data{}
	d.reset()
	cache = &Cache{db: db}
	cache.data.Store(&d)
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
//
// The cache is optimized for reads. Concurrent reads are safe
// because of the copy-on-write behavior. The txMutex just ensures
// Tx, Reset and Refresh are committed sequentially.
type Cache struct {
	txMutex     sync.Mutex
	data        atomic.Pointer[Data]
	refreshOnce sync.Once
	db          *gorm.DB
}

// Reset clears all cached data.
func (r *Cache) Reset() {
	r.txMutex.Lock()
	defer r.txMutex.Unlock()
	d := Data{}
	d.reset()
	r.data.Store(&d)
}

// Refresh reloads all data from the database.
func (r *Cache) Refresh() (err error) {
	r.txMutex.Lock()
	defer r.txMutex.Unlock()
	d := Data{}
	d.reset()
	err = d.refresh(r.db)
	if err != nil {
		return
	}
	r.data.Store(&d)
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

// GrantDeleted removes a grant and all associated tokens from the cache.
func (r *Cache) GrantDeleted(id uint) {
	_ = r.Transaction(func(tx *Tx) (_ error) {
		tx.GrantDeleted(id)
		return
	})
}

// FindTokenById returns a PAT.
func (r *Cache) FindTokenById(id uint) (m *Token, err error) {
	defer r.ensureRefreshed()
	d := r.data.Load()
	m, found := d.tokenById[id]
	if !found {
		err = &NotFound{
			Resource: "token",
			Id:       strconv.Itoa(int(id)),
		}
		return
	}
	return
}

// FindToken returns a PAT.
func (r *Cache) FindToken(token string) (m *Token, err error) {
	defer r.ensureRefreshed()
	d := r.data.Load()
	m, found := d.tokenByDigest[secret.Hash(token)]
	if !found {
		err = &NotFound{
			Resource: "token",
			Id:       token,
		}
		return
	}
	return
}

// FindSubject returns a subject.
// Tasks are not stored for performance reasons. They are found
// by matching the encoded task subject. There is nothing to be
// gained by storing them.
func (r *Cache) FindSubject(subject string) (s *Subject, err error) {
	defer r.ensureRefreshed()
	d := r.data.Load()
	user, found := d.userBySubject[subject]
	if found {
		s = &Subject{}
		var scopes []string
		scopes, err = r.FindUserScopes(user.ID)
		if err != nil {
			return
		}
		s.WithUser(user, scopes)
		return
	}
	identity, found := d.identBySubject[subject]
	if found {
		s = &Subject{}
		s.WithIdentity(identity)
		return
	}
	client, found := d.clientBySubject[subject]
	if found {
		s = &Subject{}
		s.WithClient(client)
		return
	}
	task := &Task{}
	matched := task.With(subject)
	if matched {
		s = &Subject{}
		s.WithTask(task)
		return
	}
	err = &NotFound{
		Resource: "subject",
		Id:       subject,
	}
	return
}

// FindIdentityByLogin finds and returns identities by login.
func (r *Cache) FindIdentityByLogin(login string) (m *Identity, err error) {
	defer r.ensureRefreshed()
	d := r.data.Load()
	m, found := d.identByLogin[login]
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
	defer r.ensureRefreshed()
	d := r.data.Load()
	m, found := d.roleById[id]
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
	defer r.ensureRefreshed()
	d := r.data.Load()
	m, found := d.roleByName[name]
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
	defer r.ensureRefreshed()
	d := r.data.Load()
	m, found := d.userByLogin[login]
	if !found {
		err = &NotFound{
			Resource: "user",
			Id:       login,
		}
	}
	return
}

// FindUserScopes returns the user scopes.
func (r *Cache) FindUserScopes(id uint) (scopes []string, err error) {
	defer r.ensureRefreshed()
	d := r.data.Load()
	scopes, found := d.userScopes[id]
	if !found {
		err = &NotFound{
			Resource: "user",
			Id:       strconv.Itoa(int(id)),
		}
	}
	return
}

// Begin returns a cache transaction.
// The transaction holds a changelog of cache operations.
// Changes are applied atomically on Commit().
func (r *Cache) Begin() (tx *Tx) {
	tx = &Tx{
		changes: make([]func(*Data), 0),
		cache:   r,
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

// ensureRefreshed detected the cache is stale and
// refresh (Asynchronously) as needed.
func (r *Cache) ensureRefreshed() {
	d := r.data.Load()
	if time.Since(d.refreshed) < Settings.CacheLifespan {
		return
	}
	d.refreshOnce.Do(func() {
		go func() {
			err := r.Refresh()
			if err != nil {
				Log.Error(err, "REFRESH FAILED:"+err.Error())
				r.Reset()
			}
		}()
	})
}

// Data contains cached maps.
type Data struct {
	refreshOnce sync.Once
	refreshed   time.Time
	//
	permById        map[uint]*Permission
	roleById        map[uint]*Role
	roleByName      map[string]*Role
	userById        map[uint]*User
	userBySubject   map[string]*User
	userByLogin     map[string]*User
	userScopes      map[uint][]string
	identById       map[uint]*Identity
	identBySubject  map[string]*Identity
	identByLogin    map[string]*Identity
	clientById      map[uint]*IdpClient
	clientBySubject map[string]*IdpClient
	tokenById       map[uint]*Token
	tokenByDigest   map[string]*Token
}

// reset creates new maps.
func (d *Data) reset() {
	d.permById = make(map[uint]*Permission)
	d.roleById = make(map[uint]*Role)
	d.roleByName = make(map[string]*Role)
	d.userById = make(map[uint]*User)
	d.userBySubject = make(map[string]*User)
	d.userByLogin = make(map[string]*User)
	d.userScopes = make(map[uint][]string)
	d.identById = make(map[uint]*Identity)
	d.identBySubject = make(map[string]*Identity)
	d.identByLogin = make(map[string]*Identity)
	d.clientById = make(map[uint]*IdpClient)
	d.clientBySubject = make(map[string]*IdpClient)
	d.tokenById = make(map[uint]*Token)
	d.tokenByDigest = make(map[string]*Token)
}

// clone returns cloned data.
// the refreshed timestamp is copied, but the
// refreshOnce is new.
func (d *Data) clone() *Data {
	return &Data{
		refreshOnce:     sync.Once{},
		refreshed:       d.refreshed,
		permById:        cloneMap(d.permById),
		roleById:        cloneMap(d.roleById),
		roleByName:      cloneMap(d.roleByName),
		userById:        cloneMap(d.userById),
		userBySubject:   cloneMap(d.userBySubject),
		userByLogin:     cloneMap(d.userByLogin),
		userScopes:      cloneMap(d.userScopes),
		identById:       cloneMap(d.identById),
		identBySubject:  cloneMap(d.identBySubject),
		identByLogin:    cloneMap(d.identByLogin),
		clientById:      cloneMap(d.clientById),
		clientBySubject: cloneMap(d.clientBySubject),
		tokenById:       cloneMap(d.tokenById),
		tokenByDigest:   cloneMap(d.tokenByDigest),
	}
}

// refresh cached content.
func (d *Data) refresh(db *gorm.DB) (err error) {
	d.reset()
	err = d.getPerms(db)
	if err != nil {
		return
	}
	err = d.getRoles(db)
	if err != nil {
		return
	}
	err = d.getUsers(db)
	if err != nil {
		return
	}
	err = d.getIdentities(db)
	if err != nil {
		return
	}
	err = d.getClients(db)
	if err != nil {
		return
	}
	err = d.getTokens(db)
	if err != nil {
		return
	}

	d.updateScopes()

	d.refreshed = time.Now()

	return
}

// getPerms fetches permissions from the DB and populates.
func (d *Data) getPerms(db *gorm.DB) (err error) {
	list := make([]*Permission, 0)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		d.permById[m.ID] = m
	}
	return
}

// getRoles fetches roles from the DB and populates.
func (d *Data) getRoles(db *gorm.DB) (err error) {
	list := make([]*Role, 0)
	db = db.Preload(clause.Associations)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		d.roleById[m.ID] = m
		d.roleByName[m.Name] = m
	}
	return
}

// getUsers fetches users from the DB and populates.
func (d *Data) getUsers(db *gorm.DB) (err error) {
	list := make([]*User, 0)
	db = db.Preload(clause.Associations)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		d.userById[m.ID] = m
		d.userBySubject[m.Subject] = m
		d.userByLogin[m.Login] = m
	}
	return
}

// getIdentities fetches idp identities from the DB and populates.
func (d *Data) getIdentities(db *gorm.DB) (err error) {
	list := make([]*Identity, 0)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		err = secret.Decrypt(m)
		if err == nil {
			d.identById[m.ID] = m
			d.identBySubject[m.Subject] = m
			d.identByLogin[m.Login] = m
		} else {
			return
		}
	}
	return
}

// getClients fetches idp clients from the DB and populates.
func (d *Data) getClients(db *gorm.DB) (err error) {
	list := make([]*IdpClient, 0)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		d.clientById[m.ID] = m
		d.clientBySubject[m.Subject] = m
	}
	return
}

// getTokens fetches permissions from the DB and populates.
func (d *Data) getTokens(db *gorm.DB) (err error) {
	list := []*Token{}
	db = db.Preload(clause.Associations)
	db = db.Where("kind", KindAPIKey)
	db = db.Where("expiration > ?", time.Now())
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		d.tokenById[m.ID] = m
		d.tokenByDigest[m.Digest] = m
	}
	return
}

// addTokenScopes determine token scopes and update the data.
func (d *Data) addTokenScopes(m *Token) {
	if m.Kind != KindAPIKey {
		return
	}
	var err error
	defer func() {
		if err != nil {
			Log.Info(err.Error())
		}
	}()
	// user binding.
	if m.UserID != nil {
		scopes, found := d.userScopes[*m.UserID]
		if !found {
			err = &NotFound{
				Resource: "user.scopes",
				Id:       strconv.Itoa(int(*m.UserID)),
			}
			return
		}
		m.Scopes = scopes
		return
	}
	// task binding.
	if m.TaskID != nil {
		m.Scopes = AddonScopes
		return
	}
	// IdP identity binding.
	if m.IdpIdentityID != nil {
		identity, found := d.identById[*m.IdpIdentityID]
		if !found {
			err = &NotFound{
				Resource: "identity",
				Id:       strconv.Itoa(int(*m.IdpIdentityID)),
			}
			return
		}
		m.Scopes = strings.Fields(identity.Scopes)
		return
	}
	// IdP client binding.
	if m.IdpClientID != nil {
		client, found := d.clientById[*m.IdpClientID]
		if !found {
			err = &NotFound{
				Resource: "client",
				Id:       strconv.Itoa(int(*m.IdpClientID)),
			}
			return
		}
		m.Scopes = client.GetScopes()
		return
	}
	return
}

// addUserScopes determine user scopes and add to the data.
func (d *Data) addUserScopes(m *User) {
	scopes := []string{}
	for _, r := range m.Roles {
		r, found := d.roleById[r.ID]
		if !found {
			err := &NotFound{
				Resource: "role",
				Id:       strconv.Itoa(int(r.ID)),
			}
			Log.Info(err.Error())
			continue
		}
		for _, p := range r.Permissions {
			scopes = append(scopes, p.Scope)
		}
	}
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)
	d.userScopes[m.ID] = scopes
}

// updateScopes
func (d *Data) updateScopes() {
	d.userScopes = make(map[uint][]string)
	for _, m := range d.userById {
		d.addUserScopes(m)
	}
	for _, m := range d.tokenById {
		d.addTokenScopes(m)
	}
}

// cloneMap returns a shallow clone.
func cloneMap[K comparable, V any](m map[K]V) (m2 map[K]V) {
	if m == nil {
		return nil
	}
	m2 = make(map[K]V, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return
}
