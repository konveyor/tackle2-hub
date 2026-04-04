package auth

import (
	"context"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
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
	m := &model.Token{
		TokenId:    token.ID,
		GrantId:    token.GrantID,
		ClientId:   token.ClientID,
		Subject:    token.Subject,
		Type:       string(token.Type),
		Scopes:     token.Scopes,
		Issued:     asTime(token.CreatedAtTimestamp),
		Expiration: asTime(token.ExpiresAtTimestamp),
	}
	m.UserID, err = r.getUser(token)
	if err != nil {
		return
	}
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
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
	err = secret.Decrypt(m)
	if err != nil {
		return
	}
	token = &goidc.Token{
		ID:                 m.TokenId,
		GrantID:            m.GrantId,
		ClientID:           m.ClientId,
		Subject:            m.Subject,
		Type:               goidc.TokenType(m.Type),
		Scopes:             m.Scopes,
		CreatedAtTimestamp: asInt(m.Issued),
		ExpiresAtTimestamp: asInt(m.Expiration),
	}
	return
}

// ByRefreshToken returns a grant by refresh token.
func (r *TokenManager) ByRefreshToken(token string) (m *model.Token, err error) {
	err = r.db.First(m, "refreshToken", token).Error
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
func (r *TokenManager) getUser(token *goidc.Token) (userId *uint, err error) {
	grantManager := NewGrantManager(r.db)
	grant, err := grantManager.Grant(context.TODO(), token.GrantID)
	if err != nil {
		return
	}
	if grant.Type != goidc.GrantAuthorizationCode {
		return
	}
	user := &model.User{}
	err = r.db.First(user, "uuid", token.Subject).Error
	if err != nil {
		return
	}
	userId = &user.ID
	return
}
