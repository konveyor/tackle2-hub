package auth

import (
	"context"
	"errors"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"gorm.io/gorm"
)

func NewGrantManager(db *gorm.DB) (m *GrantManager) {
	m = &GrantManager{db: db}
	return
}

type GrantManager struct {
	db *gorm.DB
}

// Save stores a grant.
func (r *GrantManager) Save(_ context.Context, grant *goidc.Grant) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.First(m, "GrantId", grant.ID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = liberr.Wrap(err)
		return
	}
	m.GrantId = grant.ID
	m.ClientId = grant.ClientID
	m.Subject = grant.Subject
	m.RefreshToken = grant.RefreshToken
	m.AuthCode = grant.AuthCode
	m.Type = string(grant.Type)
	m.Scopes = grant.Scopes
	m.Resources = grant.Resources
	m.Expiration = asTime(grant.ExpiresAtTimestamp)
	err = r.db.Save(m).Error
	return
}

// Grant returns a grant by id.
func (r *GrantManager) Grant(_ context.Context, id string) (grant *goidc.Grant, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.First(m, "grantId", id).Error
	if err != nil {
		err = notFound(err)
		return
	}
	grant, err = r.grant(m)
	if err != nil {
		return
	}
	return
}

// GrantByRefreshToken returns a grant by refresh token.
// Revocation is applied.
func (r *GrantManager) GrantByRefreshToken(_ context.Context, token string) (grant *goidc.Grant, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.First(m, "refreshToken", token).Error
	if err != nil {
		err = notFound(err)
		return
	}
	err = r.revoked(m)
	if err != nil {
		return
	}
	err = r.orphaned(m)
	if err != nil {
		return
	}
	grant, err = r.grant(m)
	return
}

// Delete a grant by id.
func (r *GrantManager) Delete(_ context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.Delete(m, "grantId", id).Error
	return
}

// DeleteByAuthCode delete a grant by auth code.
func (r *GrantManager) DeleteByAuthCode(_ context.Context, authCode string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Grant{}
	err = r.db.Delete(m, "authCode", authCode).Error
	return
}

// grant returns a goidc.Grant using
// the decrypted grant.
func (r *GrantManager) grant(m *model.Grant) (grant *goidc.Grant, err error) {
	grant = &goidc.Grant{
		ID:                 m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		RefreshToken:       m.RefreshToken,
		AuthCode:           m.AuthCode,
		Type:               goidc.GrantType(m.Type),
		Scopes:             m.Scopes,
		Resources:          m.Resources,
		ExpiresAtTimestamp: asInt(m.Expiration),
	}
	return
}

// revoked enforces token revocation.
// When the access token associated with a granted by
// refresh token has been revoked, the grant expiration
// is updated to match the revocation timestamp.
func (r *GrantManager) revoked(grant *model.Grant) (err error) {
	tokenManager := NewTokenManager(r.db)
	token, err := tokenManager.ByGrantId(grant.GrantId)
	if err != nil {
		return
	}
	if !token.Revoked.IsZero() {
		grant.Expiration = token.Revoked
		err = r.db.Save(grant).Error
	}
	return
}

// orphaned imposes grant expiration when the
// user cannot be found.
func (r *GrantManager) orphaned(grant *model.Grant) (err error) {
	if grant.Type != string(goidc.GrantAuthorizationCode) {
		return
	}
	count := int64(0)
	user := &model.User{}
	db := r.db.Model(user)
	db = db.Where("subject", grant.Subject)
	err = db.Count(&count).Error
	if err != nil {
		err = notFound(err)
		return
	}
	if count == 0 {
		grant.Expiration = time.Now().UTC()
	}
	return
}
