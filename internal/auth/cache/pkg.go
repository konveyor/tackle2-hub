package cache

import (
	"sort"
	"strconv"
	"strings"

	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
)

const (
	KindAuthCode    = "authCode"
	KindAPIKey      = api.TokenKindAPIKey
	KindAccessToken = api.TokenKindAccess
)

const (
	IdentityKindOpenid = "openid"
	IdentityKindLDAP   = "ldap"
)

// Model alias
type Model = model.Model

// User alias.
type User = model.User

// Role alias.
type Role = model.Role

// Permission alias.
type Permission = model.Permission

// Task (lightweight) model.
type Task struct {
	ID uint
}

// Login returns the (simulated) login.
// Format: task.{id}
func (m Task) Login() (s string) {
	id := strconv.FormatUint(uint64(m.ID), 10)
	s = "task." + id
	return
}

// Subject returns the task (encoded) subject.
// Format: task.0x{id}.
func (m Task) Subject() (s string) {
	id := strconv.FormatUint(uint64(m.ID), 16)
	s = "task.0x" + id
	return
}

// With populates the task.
// matched indicates the encoded subject is a task.
func (m *Task) With(subject string) (matched bool) {
	if !strings.HasPrefix(subject, "task.0x") {
		return
	}
	id := subject[7:]
	uintId, err := strconv.ParseUint(id, 16, 64)
	if err == nil {
		m.ID = uint(uintId)
		matched = true
	}
	return
}

// GetScopes returns the task scopes.
func (m Task) GetScopes() (scopes []string) {
	scopes = AddonScopes
	return
}

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
