package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strconv"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"gorm.io/gorm"
)

// NewKeyManager returns a configured key manager.
func NewKeyManager(db *gorm.DB) (m *KeyManager) {
	m = &KeyManager{db: db}
	return
}

// KeyManager manages RSA keys.
type KeyManager struct {
	db *gorm.DB
}

// KeySet returns a keyset.
// Rotation is applied.
func (r *KeyManager) KeySet() (keySet KeySet, err error) {
	var keyList []*model.RsaKey
	db := r.db.Order("id desc")
	err = db.Find(&keyList).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	created, err := r.rotate(keyList)
	if err != nil {
		return
	}
	keyList = append(created, keyList...)
	for _, m := range keyList {
		err = secret.Decrypt(m)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		b := []byte(m.PEM)
		decoded, _ := pem.Decode(b)
		var key *rsa.PrivateKey
		key, err = x509.ParsePKCS1PrivateKey(decoded.Bytes)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		jwKey := r.jwKey(m.ID, key)
		keySet.Keys = append(keySet.Keys, jwKey)
	}
	return
}

// rotate returns a new RSA key as determined
// by the rotation schedule.
func (r *KeyManager) rotate(keyList []*model.RsaKey) (created []*model.RsaKey, err error) {
	threshold := Settings.Auth.Key.Rotation
	for _, key := range keyList {
		age := time.Since(key.CreateTime)
		if age < threshold {
			return
		}
	}
	_, m := r.newKey()
	err = secret.Encrypt(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.db.Create(&m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	created = append(created, m)
	return
}

// newKey returns a new RSA key.
func (r *KeyManager) newKey() (key *rsa.PrivateKey, m *model.RsaKey) {
	m = &model.RsaKey{}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	b := x509.MarshalPKCS1PrivateKey(key)
	b = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: b,
	})
	m.PEM = string(b)
	return
}

// jwKey returns a goidc.JSONWebKey.
func (r *KeyManager) jwKey(id uint, k *rsa.PrivateKey) (k2 goidc.JSONWebKey) {
	k2.Key = strconv.Itoa(int(id))
	k2.Algorithm = "RS256"
	k2.Key = k
	return
}
