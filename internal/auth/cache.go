package auth

import (
	"sync"
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
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
	hashedSecret := hashSecret(nakedSecret)
	db := r.db.Preload(clause.Associations)
	db = db.Preload("User.Roles")
	db = db.Preload("User.Roles.Permissions")
	db = db.Where("secret", hashedSecret)
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
	return
}
