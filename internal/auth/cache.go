package auth

import (
	"errors"
	"sync"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KeyCache struct {
	db *gorm.DB
	mu sync.RWMutex
}

func (r *KeyCache) Get(nakedSecret string) (key APIKey, err error) {
	m := &model.APIKey{}
	encrypted := nakedSecret
	err = secret.Encrypt(&encrypted)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("User.Roles")
	db = db.Preload("User.Roles.Permissions")
	err = db.First(m, "secret", encrypted).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = &NotAuthenticated{
				Token: encrypted,
			}
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	key.Secret = nakedSecret
	key.Expiration = m.Expiration
	if m.UserID != nil {
		if m.User == nil {
			err = &NotAuthenticated{
				Token: encrypted,
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
				Token: encrypted,
			}
			return
		}
		switch m.Task.State {
		case task.Succeeded,
			task.Failed,
			task.Canceled:
			err = &NotAuthenticated{
				Token: encrypted,
			}
			return
		}
	}
	return
}
