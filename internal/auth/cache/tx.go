package cache

import "github.com/konveyor/tackle2-hub/shared/task"

// Tx is a cache transaction.
type Tx struct {
	cache   *Cache
	changes []func()
}

// RoleSaved updates the cache when a role is saved.
func (r *Tx) RoleSaved(m *Role) {
	r.changes = append(
		r.changes, func() {
			r.cache.roleById[m.ID] = m
			r.cache.roleByName[m.Name] = m
		})
}

// RoleDeleted removes a role from the cache.
func (r *Tx) RoleDeleted(id uint) {
	r.changes = append(
		r.changes, func() {
			m, found := r.cache.roleById[id]
			if found {
				delete(r.cache.roleById, id)
				delete(r.cache.roleByName, m.Name)
			}
		})
}

// UserSaved updates the cache when a user is saved.
func (r *Tx) UserSaved(m *User) {
	r.changes = append(
		r.changes, func() {
			r.cache.userById[m.ID] = m
			r.cache.userBySubject[m.Subject] = m
			r.cache.userByLogin[m.Login] = m
		})
}

// UserDeleted removes a user from the cache.
func (r *Tx) UserDeleted(id uint) {
	r.changes = append(
		r.changes, func() {
			m, found := r.cache.userById[id]
			if found {
				delete(r.cache.userBySubject, m.Subject)
				delete(r.cache.userByLogin, m.Login)
				delete(r.cache.userById, id)
			}
		})
}

// TaskGranted token granted.
func (r *Tx) TaskGranted(id uint) {
	r.changes = append(
		r.changes, func() {
			m := &Task{}
			m.ID = id
			m.State = task.Running
			r.cache.taskById[id] = m
		})
}

// TaskRevoked task token revoked.
func (r *Tx) TaskRevoked(id uint) {
	r.changes = append(
		r.changes, func() {
			delete(r.cache.taskById, id)
			for _, m := range r.cache.tokenById {
				if m.TaskID != nil && *m.TaskID == id {
					delete(r.cache.tokenByDigest, m.Digest)
					delete(r.cache.tokenById, m.ID)
				}
			}
		})
}

// IdentitySaved updates the cache when an identity is saved.
func (r *Tx) IdentitySaved(m *Identity) {
	r.changes = append(
		r.changes, func() {
			r.cache.identById[m.ID] = m
			r.cache.identBySubject[m.Subject] = m
			r.cache.identByLogin[m.Login] = m
		})
}

// IdentityDeleted removes an identity from the cache.
func (r *Tx) IdentityDeleted(id uint) {
	r.changes = append(
		r.changes, func() {
			m, found := r.cache.identById[id]
			if found {
				delete(r.cache.identByLogin, m.Login)
				delete(r.cache.identBySubject, m.Subject)
				delete(r.cache.identById, id)
			}
		})
}

// ClientSaved updates the cache when a client is saved.
func (r *Tx) ClientSaved(m *IdpClient) {
	r.changes = append(
		r.changes, func() {
			r.cache.clientById[m.ID] = m
			r.cache.clientBySubject[m.Subject] = m
		})
}

// ClientDeleted removes a client from the cache.
func (r *Tx) ClientDeleted(id uint) {
	r.changes = append(
		r.changes, func() {
			m, found := r.cache.clientById[id]
			if found {
				delete(r.cache.clientBySubject, m.Subject)
				delete(r.cache.clientById, id)
			}
		})
}

// TokenSaved updates the cache when a token is saved.
func (r *Tx) TokenSaved(m *Token) {
	r.changes = append(
		r.changes, func() {
			r.cache.tokenById[m.ID] = m
			r.cache.tokenByDigest[m.Digest] = m
		})
}

// TokenDeleted removes a token from the cache.
func (r *Tx) TokenDeleted(id uint) {
	r.changes = append(
		r.changes, func() {
			token, found := r.cache.tokenById[id]
			if found {
				delete(r.cache.tokenByDigest, token.Digest)
				delete(r.cache.tokenById, id)
			}
		})
}

// Commit applies all changelog operations atomically.
func (r *Tx) Commit() {
	r.cache.mutex.Lock()
	defer r.cache.mutex.Unlock()
	for _, fn := range r.changes {
		fn()
	}
	r.changes = nil
}

// Rollback discards the changelog.
// Safe to call multiple times (idempotent for defer).
func (r *Tx) Rollback() {
	r.changes = nil
}
