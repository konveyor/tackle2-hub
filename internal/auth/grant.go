package auth

import (
	"context"
	"strings"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	m := &model.Grant{
		GrantId:      grant.ID,
		ClientId:     grant.ClientID,
		Subject:      grant.Subject,
		RefreshToken: grant.RefreshToken,
		AuthCode:     grant.AuthCode,
		Type:         string(grant.Type),
		Scopes:       grant.Scopes,
		Expiration:   asTime(grant.ExpiresAtTimestamp),
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
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
	err = r.injectScopes(grant)
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
	m := &model.Grant{
		RefreshToken: token,
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.First(m, "refreshToken", m.RefreshToken).Error
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
	err = secret.Decrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	grant = &goidc.Grant{
		ID:                 m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		RefreshToken:       m.RefreshToken,
		AuthCode:           m.AuthCode,
		Type:               goidc.GrantType(m.Type),
		Scopes:             m.Scopes,
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
	token, err := tokenManager.ByRefreshToken(grant.RefreshToken)
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
	var count int64
	user := &model.User{}
	db := r.db.Model(user)
	db = db.Where("uuid", grant.Subject)
	err = r.db.Count(&count).Error
	if err != nil {
		return
	}
	if count == 0 {
		grant.Expiration = time.Now().UTC()
	}
	return
}

// injectScopes injects user scopes into the grant.
func (r *GrantManager) injectScopes(grant *goidc.Grant) (err error) {
	if grant.Type != goidc.GrantAuthorizationCode {
		return
	}
	user := &model.User{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("Roles.Permissions")
	err = db.First(user, "uuid", grant.Subject).Error
	if err != nil {
		err = notFound(err)
		return
	}
	if len(user.Roles) == 0 {
		return
	}
	unique := make(map[string]byte)
	for _, scope := range strings.Fields(grant.Scopes) {
		scope = strings.TrimSpace(scope)
		unique[scope] = 0
	}
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			unique[permission.Scope] = 0
		}
	}
	scopes := make([]string, 0, len(unique))
	for scope := range unique {
		scopes = append(scopes, scope)
	}
	grant.Scopes = strings.Join(scopes, " ")
	return
}

// tokenOptions returns the token options for the specified grant.
func (r *GrantManager) tokenOptions(
	_ context.Context,
	grant *goidc.Grant,
	client *goidc.Client) (opt goidc.TokenOptions) {
	//
	opt = goidc.NewJWTTokenOptions(goidc.RS256, 300)
	if grant.Scopes == "" {
		grant.Scopes = client.ScopeIDs
	}
	scopes := grant.Scopes
	err := r.injectScopes(grant)
	if err != nil {
		Log.Error(
			err,
			"Failed to inject permission scopes.",
			"GrantId", grant.ID,
			"Subject", grant.Subject)
		grant.Scopes = scopes
	}
	return
}
