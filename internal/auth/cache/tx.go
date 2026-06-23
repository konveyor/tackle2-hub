package cache

// Tx is a cache transaction.
type Tx struct {
	cache   *Cache
	changes []func(d *Data)
}

// RoleSaved updates the cache when a role is saved.
func (r *Tx) RoleSaved(m *Role) {
	r.changes = append(
		r.changes, func(d *Data) {
			p, found := d.roleById[m.ID]
			if found {
				delete(d.roleByName, p.Name)
			}
			d.roleById[m.ID] = m
			d.roleByName[m.Name] = m
		})
}

// RoleDeleted removes a role from the cache.
func (r *Tx) RoleDeleted(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			m, found := d.roleById[id]
			if found {
				delete(d.roleById, id)
				delete(d.roleByName, m.Name)
			}
		})
}

// UserSaved updates the cache when a user is saved.
func (r *Tx) UserSaved(m *User) {
	r.changes = append(
		r.changes, func(d *Data) {
			d.userById[m.ID] = m
			d.userBySubject[m.Subject] = m
			d.userByLogin[m.Login] = m
		})
}

// UserDeleted removes a user and all associated tokens from the cache.
func (r *Tx) UserDeleted(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			m, found := d.userById[id]
			if found {
				delete(d.userBySubject, m.Subject)
				delete(d.userByLogin, m.Login)
				delete(d.userById, id)
			}
			for _, token := range d.tokenById {
				if token.UserID != nil && *token.UserID == id {
					delete(d.tokenByDigest, token.Digest)
					delete(d.tokenById, token.ID)
				}
			}
		})
}

// TaskRevoked task token revoked.
func (r *Tx) TaskRevoked(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			for _, m := range d.tokenById {
				if m.TaskID != nil && *m.TaskID == id {
					delete(d.tokenByDigest, m.Digest)
					delete(d.tokenById, m.ID)
				}
			}
		})
}

// IdentitySaved updates the cache when an identity is saved.
func (r *Tx) IdentitySaved(m *Identity) {
	r.changes = append(
		r.changes, func(d *Data) {
			d.identById[m.ID] = m
			d.identBySubject[m.Subject] = m
			d.identByLogin[m.Login] = m
		})
}

// IdentityDeleted removes an identity and all associated tokens from the cache.
func (r *Tx) IdentityDeleted(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			m, found := d.identById[id]
			if found {
				delete(d.identByLogin, m.Login)
				delete(d.identBySubject, m.Subject)
				delete(d.identById, id)
			}
			for _, token := range d.tokenById {
				if token.IdpIdentityID != nil && *token.IdpIdentityID == id {
					delete(d.tokenByDigest, token.Digest)
					delete(d.tokenById, token.ID)
				}
			}
		})
}

// ClientSaved updates the cache when a client is saved.
func (r *Tx) ClientSaved(m *IdpClient) {
	r.changes = append(
		r.changes, func(d *Data) {
			d.clientById[m.ID] = m
			d.clientBySubject[m.Subject] = m
		})
}

// ClientDeleted removes a client and all associated tokens from the cache.
func (r *Tx) ClientDeleted(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			m, found := d.clientById[id]
			if found {
				delete(d.clientBySubject, m.Subject)
				delete(d.clientById, id)
			}
			for _, token := range d.tokenById {
				if token.IdpClientID != nil && *token.IdpClientID == id {
					delete(d.tokenByDigest, token.Digest)
					delete(d.tokenById, token.ID)
				}
			}
		})
}

// TokenSaved updates the cache when a token is saved.
func (r *Tx) TokenSaved(m *Token) {
	r.changes = append(
		r.changes, func(d *Data) {
			d.tokenById[m.ID] = m
			d.tokenByDigest[m.Digest] = m
		})
}

// TokenDeleted removes a token from the cache.
func (r *Tx) TokenDeleted(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			token, found := d.tokenById[id]
			if found {
				delete(d.tokenByDigest, token.Digest)
				delete(d.tokenById, id)
			}
		})
}

// GrantDeleted removes a grant and all associated tokens from the cache.
func (r *Tx) GrantDeleted(id uint) {
	r.changes = append(
		r.changes, func(d *Data) {
			for _, m := range d.tokenById {
				if m.GrantID != nil && *m.GrantID == id {
					delete(d.tokenByDigest, m.Digest)
					delete(d.tokenById, m.ID)
				}
			}
		})
}

// changed notification that Data has changed.
func (r *Tx) changed(d *Data) {
	d.updateScopes()
}

// Commit applies all changelog operations atomically.
func (r *Tx) Commit() {
	r.cache.txMutex.Lock()
	defer r.cache.txMutex.Unlock()
	d := r.cache.data.Load()
	clone := d.clone()
	for _, fn := range r.changes {
		fn(clone)
	}
	r.changed(clone)
	r.cache.data.Store(clone)
	r.changes = nil
}

// Rollback discards the changelog.
// Safe to call multiple times (idempotent for defer).
func (r *Tx) Rollback() {
	r.changes = nil
}
