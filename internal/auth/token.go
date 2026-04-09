package auth

import (
	"context"
	"errors"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"gorm.io/gorm"
)

// NewTokenManager returns a token manager.
func NewTokenManager(db *gorm.DB) (m *TokenManager) {
	m = &TokenManager{db: db}
	return
}

// TokenManager manages tokens.
type TokenManager struct {
	db *gorm.DB
}

// Save stores a token.
func (r *TokenManager) Save(ctx context.Context, token *goidc.Token) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.First(m, "TokenId", token.ID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = liberr.Wrap(err)
		return
	}
	m.TokenId = token.ID
	m.GrantId = token.GrantID
	m.ClientId = token.ClientID
	m.Subject = token.Subject
	m.Type = string(token.Type)
	m.Scopes = token.Scopes
	m.Resources = token.Resources
	m.Issued = asTime(token.CreatedAtTimestamp)
	m.Expiration = asTime(token.ExpiresAtTimestamp)
	m.UserID, err = r.getUser(token)
	if err != nil {
		return
	}
	err = r.db.Save(m).Error
	return
}

// Token returns a token by id.
func (r *TokenManager) Token(ctx context.Context, id string) (token *goidc.Token, err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.First(m, "tokenId", id).Error
	if err != nil {
		err = notFound(err)
		return
	}
	token = &goidc.Token{
		ID:                 m.TokenId,
		GrantID:            m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		Type:               goidc.TokenType(m.Type),
		Scopes:             m.Scopes,
		Resources:          m.Resources,
		CreatedAtTimestamp: asInt(m.Issued),
		ExpiresAtTimestamp: asInt(m.Expiration),
	}
	return
}

// ByGrantId returns a grant by refresh token.
func (r *TokenManager) ByGrantId(grantId string) (m *model.Token, err error) {
	m = &model.Token{}
	err = r.db.First(m, "grantId", grantId).Error
	if err != nil {
		err = notFound(err)
		return
	}
	return
}

// Delete grant by id.
func (r *TokenManager) Delete(ctx context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.Delete(m, "tokenId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// DeleteByGrantID by grant by id.
func (r *TokenManager) DeleteByGrantID(_ context.Context, id string) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	m := &model.Token{}
	err = r.db.Delete(m, "grantId", id).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// GetUser returns the user ID as appropriate.
func (r *TokenManager) getUser(token *goidc.Token) (userid *uint, err error) {
	grantManager := NewGrantManager(r.db)
	grant, err := grantManager.Grant(context.TODO(), token.GrantID)
	if err != nil {
		return
	}
	if grant.Type != goidc.GrantAuthorizationCode {
		return
	}
	user := &model.User{}
	err = r.db.First(user, "subject", token.Subject).Error
	if err != nil {
		return
	}
	userid = &user.ID
	return
}
