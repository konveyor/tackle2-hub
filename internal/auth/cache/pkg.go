package cache

import (
	"sort"

	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
)

const (
	KindAccessToken = "access"
	KindAuthCode    = "authCode"
	KindAPIKey      = "api-key"
)

const (
	IdentityKindOpenid = "openid"
	IdentityKindLDAP   = "ldap"
)

// Model alias
type Model = model.Model

// User alias.
type User model.User

// GetScopes returns the user's scopes.
func (m *User) GetScopes(cache *Cache) (scopes []string) {
	for _, r := range m.Roles {
		role, nErr := cache.FindRoleById(r.ID)
		if nErr != nil {
			continue
		}
		for _, scope := range role.GetScopes() {
			scopes = append(scopes, scope)
		}
	}
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)
	return
}

// Role alias.
type Role model.Role

// ScopeNames returns the roles scopes.
func (m *Role) GetScopes() (scopes []string) {
	for _, p := range m.Permissions {
		scopes = append(scopes, p.Scope)
	}
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)
	return
}

// Permission alias.
type Permission = model.Permission

// Task alias.
type Task = model.Task

// Identity alias.
type Identity = model.IdpIdentity

// IdpClient alias.
type IdpClient model.IdpClient

// GetScopes returns the client's scopes.
func (m *IdpClient) GetScopes() (scopes []string) {
	scopes = m.Scopes
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)
	return
}

// With populates self with the settings client.
func (m *IdpClient) With(client *as.IdpClient) {
	m.ID = client.ID
	m.ClientId = client.ClientId
	m.Secret = client.Secret
	m.ApplicationType = client.ApplicationType
	m.Grants = client.Grants
	m.RedirectURIs = client.RedirectURIs
	m.Scopes = client.Scopes
}

// Grant alias.
type Grant = model.Grant

// Token alias.
type Token struct {
	model.Token
	Secret string `gorm:"-"`
}

// uniqueStrings returns a list with duplicates removed.
func uniqueStrings(items []string) (result []string) {
	seen := make(map[string]bool, len(items))
	result = make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return
}
