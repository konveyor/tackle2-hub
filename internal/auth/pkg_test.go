package auth

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TestUserGrant tests creating and authenticating with user tokens.
func TestUserGrant(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user
	user := &model.User{
		Subject:  "test-uuid-123",
		Userid:   "testuser",
		Password: secret.HashPassword("testpassword"),
		Email:    "test@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test creating token with valid subject
	token, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.Secret).NotTo(BeEmpty())

	// Verify the token was created in the database
	var keyCount int64
	err = db.Model(&model.Token{}).Count(&keyCount).Error
	g.Expect(err).To(BeNil())
	g.Expect(keyCount).To(Equal(int64(1)))

	// Verify the digest in the database matches
	err = db.First(&token).Error
	g.Expect(err).To(BeNil())

	// Test creating token with non-existent subject
	_, err = provider.NewPAT("non-existent-subject", 24*time.Hour)
	g.Expect(err).NotTo(BeNil())

	// Test authenticating with the token
	request := &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify token claims
	claims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(claims[ClaimSub]).To(Equal("test-uuid-123"))

	// Test that expired keys are rejected
	expiredSecret := "expired-secret-key"
	expiredKey := &model.Token{
		UserID:     &user.ID,
		Expiration: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err = db.Create(expiredKey).Error
	g.Expect(err).To(BeNil())

	request = &Request{}
	request.With("Bearer " + expiredSecret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestTaskKey tests creating and authenticating with task tokens.
func TestTaskGrant(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task in Running state (required for cache to load it)
	task := &model.Task{
		Name:  "test-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test creating task token
	token, err := provider.NewTaskToken(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(token.Secret).NotTo(BeEmpty())

	// Test authenticating with the task token
	request := &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify the digest in the database matches
	err = db.First(&token).Error
	g.Expect(err).To(BeNil())

	// Test creating key for non-existent task
	_, err = provider.NewTaskToken(9999)
	g.Expect(err).NotTo(BeNil())
}

// TestJWTAuthentication tests authenticating with JWT tokens using HMAC signing.
func TestJWTAuthentication(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Use HMAC signing with the configured token key for testing
	signingKey := []byte(Settings.Auth.Token.Key)

	// Create a valid JWT token
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims[ClaimSub] = "user-123"
	claims[ClaimScope] = "openid profile email"
	claims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	claims[ClaimIss] = "test-issuer"
	claims[ClaimAud] = "test-audience"

	tokenString, err := token.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	// Test authenticating with valid JWT
	request := &Request{}
	request.With("Bearer " + tokenString)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify claims
	jwtClaims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(jwtClaims[ClaimSub]).To(Equal("user-123"))
	g.Expect(jwtClaims[ClaimScope]).To(Equal("openid profile email"))

	// Test with expired token
	expiredToken := jwt.New(jwt.SigningMethodHS512)
	expiredClaims := expiredToken.Claims.(jwt.MapClaims)
	expiredClaims[ClaimSub] = "user-123"
	expiredClaims[ClaimScope] = "openid profile"
	expiredClaims[ClaimExp] = float64(time.Now().Add(-1 * time.Hour).Unix()) // Expired
	expiredClaims[ClaimIss] = "test-issuer"

	expiredTokenString, err := expiredToken.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	request = &Request{}
	request.With("Bearer " + expiredTokenString)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("expired"))

	// Test with missing sub claim
	noSubToken := jwt.New(jwt.SigningMethodHS512)
	noSubClaims := noSubToken.Claims.(jwt.MapClaims)
	noSubClaims[ClaimScope] = "openid"
	noSubClaims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())

	noSubTokenString, err := noSubToken.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	request = &Request{}
	request.With("Bearer " + noSubTokenString)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("User not specified"))

	// Test with missing scope claim
	noScopeToken := jwt.New(jwt.SigningMethodHS512)
	noScopeClaims := noScopeToken.Claims.(jwt.MapClaims)
	noScopeClaims[ClaimSub] = "user-123"
	noScopeClaims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())

	noScopeTokenString, err := noScopeToken.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	request = &Request{}
	request.With("Bearer " + noScopeTokenString)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("Scope not specified"))
}

// TestUserExtraction tests extracting user from token.
func TestUserExtraction(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create a token with claims
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims[ClaimSub] = "test-user-456"
	claims[ClaimScope] = "openid profile"

	user := provider.User(token)
	g.Expect(user).To(Equal("test-user-456"))

	// Test with missing sub claim
	tokenNoSub := jwt.New(jwt.SigningMethodHS256)
	tokenNoSub.Claims.(jwt.MapClaims)[ClaimScope] = "openid"

	user = provider.User(tokenNoSub)
	g.Expect(user).To(BeEmpty())
}

// TestScopesExtraction tests extracting scopes from token.
func TestScopesExtraction(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create a token with multiple scopes
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims[ClaimSub] = "test-user"
	claims[ClaimScope] = "applications:read applications:write tags:read"

	scopes := provider.Scopes(token)
	g.Expect(scopes).To(HaveLen(3))

	// Verify each scope
	scopeStrings := make([]string, len(scopes))
	for i, s := range scopes {
		scopeStrings[i] = s.String()
	}
	g.Expect(scopeStrings).To(ContainElement("applications:read"))
	g.Expect(scopeStrings).To(ContainElement("applications:write"))
	g.Expect(scopeStrings).To(ContainElement("tags:read"))

	// Test with empty scope
	tokenNoScope := jwt.New(jwt.SigningMethodHS256)
	tokenNoScope.Claims.(jwt.MapClaims)[ClaimSub] = "test-user"

	scopes = provider.Scopes(tokenNoScope)
	g.Expect(scopes).To(BeEmpty())
}

// TestInvalidBearerToken tests handling of malformed bearer tokens.
func TestInvalidBearerToken(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test missing "Bearer" prefix
	request := &Request{}
	request.With("invalid-token-without-bearer")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test empty token
	request = &Request{}
	request.With("")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test malformed bearer token (missing token value)
	request = &Request{}
	request.With("Bearer")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test invalid JWT
	request = &Request{}
	request.With("Bearer invalid.jwt.token")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestBaseScopeMatching tests scope matching with wildcards and exact matches.
func TestBaseScopeMatching(t *testing.T) {
	g := NewGomegaWithT(t)

	// Wildcard scope matches everything
	scope := &BaseScope{Resource: "*", Method: "*"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("tags", "POST")).To(BeTrue())

	// Resource wildcard matches any method for that resource
	scope = &BaseScope{Resource: "applications", Method: "*"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeTrue())
	g.Expect(scope.Match("tags", "GET")).To(BeFalse())

	// Method wildcard matches that method for any resource
	scope = &BaseScope{Resource: "*", Method: "GET"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("tags", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeFalse())

	// Exact match
	scope = &BaseScope{Resource: "applications", Method: "GET"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeFalse())
	g.Expect(scope.Match("tags", "GET")).To(BeFalse())

	// Case insensitive
	scope = &BaseScope{Resource: "Applications", Method: "get"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
}

// TestBaseScopeParsing tests parsing scope strings.
func TestBaseScopeParsing(t *testing.T) {
	g := NewGomegaWithT(t)

	scope := &BaseScope{}
	scope.With("applications:read")
	g.Expect(scope.Resource).To(Equal("applications"))
	g.Expect(scope.Method).To(Equal("read"))

	scope = &BaseScope{}
	scope.With("*:*")
	g.Expect(scope.Resource).To(Equal("*"))
	g.Expect(scope.Method).To(Equal("*"))

	// Test String() roundtrip
	scope = &BaseScope{Resource: "tags", Method: "write"}
	g.Expect(scope.String()).To(Equal("tags:write"))
}

// TestKeyCacheWithTaskStates tests that keys for terminal tasks are rejected.
func TestKeyCacheWithTaskStates(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task in running state
	task := &model.Task{
		Name:  "test-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token for running task - should work
	key, err := provider.NewTaskToken(task.ID)
	g.Expect(err).To(BeNil())

	// Authenticate with key - should work
	request := &Request{}
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Update task to Succeeded - key should now be rejected
	provider.cache.Reset()
	db.Model(task).Update("State", "Succeeded")
	request = &Request{}
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test with Failed state
	provider.cache.Reset()
	db.Model(task).Update("State", "Failed")
	request = &Request{}
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test with Canceled state
	provider.cache.Reset()
	db.Model(task).Update("State", "Canceled")
	request = &Request{}
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestRequestPermit tests the complete authentication and authorization flow.
func TestRequestPermit(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create user with specific permissions
	user := &model.User{
		Subject:  "user-123",
		Userid:   "testuser",
		Password: secret.HashPassword("password"),
		Email:    "test@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Create role with permissions
	perm := &model.Permission{
		Name:  "Read Applications",
		Scope: "applications:GET",
	}
	err = db.Create(perm).Error
	g.Expect(err).To(BeNil())

	role := &model.Role{
		Name: "ApplicationReader",
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	err = db.Model(role).Association("Permissions").Append(perm)
	g.Expect(err).To(BeNil())

	err = db.Model(user).Association("Roles").Append(role)
	g.Expect(err).To(BeNil())

	// Set up provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	IdP = provider

	// Create token
	key, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Test authenticated and authorized (matching scope)
	request := &Request{
		Scope:  "applications",
		Method: "GET",
		DB:     db,
	}
	request.With("Bearer " + key.Secret)
	result, err := request.Permit()
	g.Expect(err).To(BeNil())
	g.Expect(result.Authenticated).To(BeTrue())
	g.Expect(result.Authorized).To(BeTrue())
	g.Expect(result.User).To(Equal("user-123"))

	// Test authenticated but not authorized (wrong method)
	request = &Request{
		Scope:  "applications",
		Method: "POST",
		DB:     db,
	}
	request.With("Bearer " + key.Secret)
	result, err = request.Permit()
	g.Expect(err).To(BeNil())
	g.Expect(result.Authenticated).To(BeTrue())
	g.Expect(result.Authorized).To(BeFalse())

	// Test not authenticated (invalid token)
	request = &Request{
		Scope:  "applications",
		Method: "GET",
		DB:     db,
	}
	request.With("Bearer invalid-token")
	result, err = request.Permit()
	g.Expect(err).To(BeNil())
	g.Expect(result.Authenticated).To(BeFalse())
	g.Expect(result.Authorized).To(BeFalse())
}

// TestNoAuthProvider tests the NoAuth provider fallback behavior.
func TestNoAuthProvider(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())
	builtin, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	provider := NewNoAuth(builtin)

	// Authenticate always succeeds (returns nil token, nil error)
	request := &Request{}
	request.With("Bearer any-token")
	token, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Scopes returns wildcard
	scopes := provider.Scopes(token)
	g.Expect(scopes).To(HaveLen(1))
	scope := scopes[0]
	g.Expect(scope.Match("anything", "GET")).To(BeTrue())
	g.Expect(scope.Match("anything", "POST")).To(BeTrue())

	// User returns fixed admin user
	userid := provider.User(token)
	g.Expect(userid).To(Equal("admin.noauth"))

	// Create test user.
	user := &model.User{
		Subject:  "noauth-test-subject",
		Userid:   "testuser",
		Password: secret.HashPassword("password"),
		Email:    "noauth@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		db.Delete(user)
	})

	key, err := provider.NewPAT(user.Subject, time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).ToNot(BeEmpty())

	task := &model.Task{}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		db.Delete(task)
	})

	key, err = provider.NewTaskToken(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).ToNot(BeEmpty())
}

// TestNotAuthenticatedError tests NotAuthenticated error type.
func TestNotAuthenticatedError(t *testing.T) {
	g := NewGomegaWithT(t)

	err := &NotAuthenticated{Token: "test-token"}
	g.Expect(err.Error()).To(ContainSubstring("test-token"))
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test Is() method
	var target *NotAuthenticated
	g.Expect(errors.As(err, &target)).To(BeTrue())
	g.Expect(errors.Is(err, &NotAuthenticated{})).To(BeTrue())
}

// TestNotValidError tests NotValid error type.
func TestNotValidError(t *testing.T) {
	g := NewGomegaWithT(t)

	err := &NotValid{
		Token:  "test-token",
		Reason: "expired",
	}
	g.Expect(err.Error()).To(ContainSubstring("test-token"))
	g.Expect(err.Error()).To(ContainSubstring("expired"))
	g.Expect(err.Error()).To(ContainSubstring("not-valid"))

	// Test Is() method
	var target *NotValid
	g.Expect(errors.As(err, &target)).To(BeTrue())
	g.Expect(errors.Is(err, &NotValid{})).To(BeTrue())
}

// TestCacheTokenDelete tests that TokenDeleted removes tokens from both cache indexes.
func TestCacheTokenDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user
	user := &model.User{
		Subject:  "cache-delete-user",
		Userid:   "cachedeleteuser",
		Password: secret.HashPassword("password"),
		Email:    "cachedelete@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Create provider with cache
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token
	token, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.ID).NotTo(Equal(uint(0)))

	// Verify token is in cache by authenticating
	request := &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Delete token from DB first to prevent refresh from reloading it
	err = db.Delete(&model.Token{}, token.ID).Error
	g.Expect(err).To(BeNil())

	// Delete token using cache method
	provider.cache.TokenDeleted(token.ID)

	// Verify token is removed from both indexes by checking authentication fails
	request = &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))
}

// TestBuiltinDelete tests the Builtin Delete method removes key from cache and DB.
func TestBuiltinRevoke(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user with permissions
	user := &model.User{
		Subject:  "builtin-delete-user",
		Userid:   "builtindeleteuser",
		Password: secret.HashPassword("password"),
		Email:    "builtindelete@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	perm := &model.Permission{
		Name:  "Admin",
		Scope: "*:*",
	}
	err = db.Create(perm).Error
	g.Expect(err).To(BeNil())

	role := &model.Role{Name: "Admin"}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	err = db.Model(role).Association("Permissions").Append(perm)
	g.Expect(err).To(BeNil())

	err = db.Model(user).Association("Roles").Append(role)
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token
	token, err := provider.NewPAT(user.Subject, 1*time.Hour)
	g.Expect(err).To(BeNil())

	// Populate cache by authenticating
	request := &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Verify token exists in DB
	err = db.First(&token).Error
	g.Expect(err).To(BeNil())

	// Delete using provider Delete method
	err = provider.Revoke(token.ID)
	g.Expect(err).To(BeNil())

	// Verify token is removed from DB
	err = db.First(&token).Error
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.Is(err, gorm.ErrRecordNotFound)).To(BeTrue())

	// Verify token is removed from cache - authentication should fail
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))
}

func TestLDAP(t *testing.T) {
	g := NewGomegaWithT(t)

	ldap := &LDAP{
		URL:      "ldap://f35a.redhat.com:389",
		OU:       "people",
		BaseDN:   "dc=f35a,dc=redhat,dc=com",
		Userid:   "jsmith",
		Password: "dog8code",
	}
	groups, err := ldap.Authenticate()
	g.Expect(err).To(BeNil())
	t.Log(groups)
}

// TestCacheAutoRefresh tests automatic cache refresh on miss and expiry.
func TestCacheAutoRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create user and token
	user := &model.User{
		Subject:  "cache-refresh-user",
		Userid:   "cacherefreshuser",
		Password: secret.HashPassword("password"),
		Email:    "cacherefresh@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	token, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Verify token is in cache
	request := &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Delete token from DB but not cache
	err = db.Delete(&model.Token{}, token.ID).Error
	g.Expect(err).To(BeNil())

	// Create new token in DB (not in cache yet)
	newUser := &model.User{
		Subject:  "new-user",
		Userid:   "newuser",
		Password: secret.HashPassword("password"),
		Email:    "newuser@example.com",
	}
	err = db.Create(newUser).Error
	g.Expect(err).To(BeNil())

	newToken := Token{
		Token: model.Token{
			Kind:       KindAPIKey,
			AuthId:     "new-auth-id",
			Digest:     secret.Hash("new-secret-token"),
			Expiration: time.Now().Add(24 * time.Hour),
			UserID:     &newUser.ID,
		},
		Secret: "new-secret-token",
	}
	err = db.Create(&newToken.Token).Error
	g.Expect(err).To(BeNil())

	// Trigger refresh by requesting token not in cache (miss-based refresh)
	request = &Request{}
	request.With("Bearer " + newToken.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil()) // Should succeed after auto-refresh

	// Verify old deleted token is now gone from cache (refresh removed it)
	request = &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestCacheTimeBasedRefresh tests time-based cache expiration.
func TestCacheTimeBasedRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Save original cache lifespan and restore after test
	originalLifespan := Settings.CacheLifespan
	defer func() {
		Settings.CacheLifespan = originalLifespan
	}()

	// Set very short cache lifespan
	Settings.CacheLifespan = 100 * time.Millisecond

	user := &model.User{
		Subject:  "time-refresh-user",
		Userid:   "timerefreshuser",
		Password: secret.HashPassword("password"),
		Email:    "timerefresh@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	token, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Authenticate successfully (cache is fresh)
	request := &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Create new token while cache is stale
	newUser := &model.User{
		Subject:  "new-time-user",
		Userid:   "newtimeuser",
		Password: secret.HashPassword("password"),
		Email:    "newtime@example.com",
	}
	err = db.Create(newUser).Error
	g.Expect(err).To(BeNil())

	newToken := Token{
		Token: model.Token{
			Kind:       KindAPIKey,
			AuthId:     "time-auth-id",
			Digest:     secret.Hash("time-secret-token"),
			Expiration: time.Now().Add(24 * time.Hour),
			UserID:     &newUser.ID,
		},
		Secret: "time-secret-token",
	}
	err = db.Create(&newToken.Token).Error
	g.Expect(err).To(BeNil())

	// Authenticate with new token - should trigger time-based refresh
	request = &Request{}
	request.With("Bearer " + newToken.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())
}

// TestIdpIdentityTokenBinding tests token binding to IdP identities.
func TestIdpIdentityTokenBinding(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create IdP identity
	identity := &Identity{
		Issuer:       "https://idp.example.com",
		Subject:      "idp-user-123",
		RefreshToken: "refresh-token",
		Expiration:   time.Now().Add(24 * time.Hour),
		Scopes:       "openid profile email",
		Userid:       "idpuser",
		Email:        "idpuser@example.com",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token bound to IdP identity
	token, err := provider.NewPAT(identity.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.IdpIdentityID).NotTo(BeNil())
	g.Expect(*token.IdpIdentityID).To(Equal(identity.ID))

	// Authenticate with IdP-bound token
	request := &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify claims
	claims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(claims[ClaimSub]).To(Equal("idp-user-123"))
	g.Expect(claims[ClaimScope]).To(Equal("openid profile email"))
}

// TestCacheEntityUpdates tests all Saved/Deleted methods.
func TestCacheEntityUpdates(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	// Test RoleSaved/RoleDeleted
	role := &Role{
		Model: model.Model{ID: 100},
		Name:  "TestRole",
	}
	cache.RoleSaved(role)
	cache.mutex.RLock()
	_, found := cache.roleById[100]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeTrue())

	cache.RoleDeleted(100)
	cache.mutex.RLock()
	_, found = cache.roleById[100]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeFalse())

	// Test UserSaved/UserDeleted
	user := &User{
		Model:   model.Model{ID: 200},
		Subject: "test-user",
	}
	cache.UserSaved(user)
	cache.mutex.RLock()
	_, found = cache.userById[200]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeTrue())

	cache.UserDeleted(200)
	cache.mutex.RLock()
	_, found = cache.userById[200]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeFalse())

	// Test TaskSaved/TaskDeleted
	task := &Task{
		Model: model.Model{ID: 300},
		Name:  "test-task",
		State: "Running",
	}
	cache.TaskSaved(task)
	cache.mutex.RLock()
	_, found = cache.taskById[300]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeTrue())

	cache.TaskDeleted(300)
	cache.mutex.RLock()
	_, found = cache.taskById[300]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeFalse())

	// Test IdentitySaved/IdentityDeleted
	identity := &Identity{
		Model:   model.Model{ID: 400},
		Issuer:  "https://idp.example.com",
		Subject: "idp-subject",
	}
	cache.IdentitySaved(identity)
	cache.mutex.RLock()
	_, found = cache.identById[400]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeTrue())

	cache.IdentityDeleted(400)
	cache.mutex.RLock()
	_, found = cache.identById[400]
	cache.mutex.RUnlock()
	g.Expect(found).To(BeFalse())
}

// TestCacheInconsistency tests error paths when referenced entities are missing.
func TestCacheInconsistency(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	// Create token referencing non-existent user
	userID := uint(9999)
	userTokenSecret := "inconsistent-user-token"
	userTokenDigest := secret.Hash(userTokenSecret)
	userToken := &Token{
		Token: model.Token{
			Model:  model.Model{ID: 1},
			UserID: &userID,
			Digest: userTokenDigest,
		},
	}
	cache.mutex.Lock()
	cache.tokenByDigest[userTokenDigest] = userToken
	cache.tokenById[1] = userToken
	cache.mutex.Unlock()

	_, err = cache.getToken(userTokenSecret)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("user"))

	// Create token referencing non-existent task
	taskID := uint(8888)
	taskTokenSecret := "inconsistent-task-token"
	taskTokenDigest := secret.Hash(taskTokenSecret)
	taskToken := &Token{
		Token: model.Token{
			Model:  model.Model{ID: 2},
			TaskID: &taskID,
			Digest: taskTokenDigest,
		},
	}
	cache.mutex.Lock()
	cache.tokenByDigest[taskTokenDigest] = taskToken
	cache.tokenById[2] = taskToken
	cache.mutex.Unlock()

	_, err = cache.getToken(taskTokenSecret)
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("task"))

	// Create token referencing non-existent identity
	identityID := uint(7777)
	identityTokenSecret := "inconsistent-identity-token"
	identityTokenDigest := secret.Hash(identityTokenSecret)
	identityToken := &Token{
		Token: model.Token{
			Model:         model.Model{ID: 3},
			IdpIdentityID: &identityID,
			Digest:        identityTokenDigest,
		},
	}
	cache.mutex.Lock()
	cache.tokenByDigest[identityTokenDigest] = identityToken
	cache.tokenById[3] = identityToken
	cache.mutex.Unlock()

	_, err = cache.getToken(identityTokenSecret)
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("identity"))
}

// TestCacheConcurrency tests concurrent access to the cache.
func TestCacheConcurrency(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create multiple users and tokens
	users := make([]*model.User, 10)
	tokens := make([]Token, 10)
	for i := 0; i < 10; i++ {
		users[i] = &model.User{
			Subject:  fmt.Sprintf("concurrent-user-%d", i),
			Userid:   fmt.Sprintf("concurrentuser%d", i),
			Password: secret.HashPassword("password"),
			Email:    fmt.Sprintf("concurrent%d@example.com", i),
		}
		err = db.Create(users[i]).Error
		g.Expect(err).To(BeNil())
	}

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	for i := 0; i < 10; i++ {
		tokens[i], err = provider.NewPAT(users[i].Subject, 24*time.Hour)
		g.Expect(err).To(BeNil())
	}

	// Launch concurrent GetToken operations
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tokenIdx := idx % 10
			request := &Request{}
			request.With("Bearer " + tokens[tokenIdx].Secret)
			_, err := provider.Authenticate(request)
			g.Expect(err).To(BeNil())
		}(i)
	}

	// Launch concurrent Refresh operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := provider.cache.Refresh()
			g.Expect(err).To(BeNil())
		}()
	}

	// Launch concurrent TokenSaved operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			provider.cache.TokenSaved(&tokens[idx%10])
		}(i)
	}

	wg.Wait()
}

// TestScopeCalculation tests edge cases in scope calculation.
func TestScopeCalculation(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Test 1: User with no roles
	userNoRoles := &model.User{
		Subject:  "user-no-roles",
		Userid:   "usernoroles",
		Password: secret.HashPassword("password"),
		Email:    "noroles@example.com",
	}
	err = db.Create(userNoRoles).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	token, err := provider.NewPAT(userNoRoles.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	request := &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	claims := jwToken.Claims.(jwt.MapClaims)
	scopes := claims[ClaimScope].(string)
	g.Expect(scopes).To(BeEmpty())

	// Test 2: Role with no permissions
	roleNoPerm := &model.Role{
		Name: "EmptyRole",
	}
	err = db.Create(roleNoPerm).Error
	g.Expect(err).To(BeNil())

	userEmptyRole := &model.User{
		Subject:  "user-empty-role",
		Userid:   "useremptyrole",
		Password: secret.HashPassword("password"),
		Email:    "emptyrole@example.com",
	}
	err = db.Create(userEmptyRole).Error
	g.Expect(err).To(BeNil())

	err = db.Model(userEmptyRole).Association("Roles").Append(roleNoPerm)
	g.Expect(err).To(BeNil())

	// Refresh cache to load new user
	err = provider.cache.Refresh()
	g.Expect(err).To(BeNil())

	token, err = provider.NewPAT(userEmptyRole.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	request = &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	claims = jwToken.Claims.(jwt.MapClaims)
	scopes = claims[ClaimScope].(string)
	g.Expect(scopes).To(BeEmpty())

	// Test 3: Multiple roles with overlapping permissions (deduplication)
	perm1 := &model.Permission{
		Name:  "Read",
		Scope: "apps:GET",
	}
	perm2 := &model.Permission{
		Name:  "Write",
		Scope: "apps:POST",
	}
	err = db.Create(perm1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(perm2).Error
	g.Expect(err).To(BeNil())

	role1 := &model.Role{Name: "Reader"}
	role2 := &model.Role{Name: "Writer"}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(role2).Error
	g.Expect(err).To(BeNil())

	err = db.Model(role1).Association("Permissions").Append(perm1)
	g.Expect(err).To(BeNil())
	err = db.Model(role2).Association("Permissions").Append(perm1, perm2) // perm1 overlaps
	g.Expect(err).To(BeNil())

	userMultiRole := &model.User{
		Subject:  "user-multi-role",
		Userid:   "usermultirole",
		Password: secret.HashPassword("password"),
		Email:    "multirole@example.com",
	}
	err = db.Create(userMultiRole).Error
	g.Expect(err).To(BeNil())

	err = db.Model(userMultiRole).Association("Roles").Append(role1, role2)
	g.Expect(err).To(BeNil())

	// Refresh cache
	err = provider.cache.Refresh()
	g.Expect(err).To(BeNil())

	token, err = provider.NewPAT(userMultiRole.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	request = &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	claims = jwToken.Claims.(jwt.MapClaims)
	scopes = claims[ClaimScope].(string)
	// Should have both scopes, deduplicated
	g.Expect(scopes).To(ContainSubstring("apps:GET"))
	g.Expect(scopes).To(ContainSubstring("apps:POST"))
}

// TestTokenBindingEdgeCases tests edge cases in token bindings.
func TestTokenBindingEdgeCases(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	// Test 1: Token with no bindings (all foreign keys nil)
	unboundTokenDigest := secret.Hash("unbound-token")
	cache.mutex.Lock()
	cache.tokenByDigest[unboundTokenDigest] = &Token{
		Token: model.Token{
			Model:         model.Model{ID: 999},
			UserID:        nil,
			TaskID:        nil,
			IdpIdentityID: nil,
		},
	}
	cache.mutex.Unlock()

	m, err := cache.GetToken("unbound-token")
	g.Expect(err).To(BeNil())
	g.Expect(m.Subject).To(BeEmpty())
	g.Expect(m.Scopes).To(BeEmpty())

	// Test 2: Pending task token
	pendingTask := &model.Task{
		Name:  "pending-task",
		State: "Pending",
	}
	err = db.Create(pendingTask).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load pending task
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	token, err := provider.NewTaskToken(pendingTask.ID)
	g.Expect(err).To(BeNil())

	request := &Request{}
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify task subject format
	claims := jwToken.Claims.(jwt.MapClaims)
	subject := claims[ClaimSub].(string)
	g.Expect(subject).To(ContainSubstring("task:"))
}

// TestManualCacheRefresh tests explicit Refresh() and Reset() calls.
func TestManualCacheRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	user := &model.User{
		Subject:  "manual-refresh-user",
		Userid:   "manualrefreshuser",
		Password: secret.HashPassword("password"),
		Email:    "manualrefresh@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token
	token, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Verify in cache
	request := &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Call Reset - should clear cache
	provider.cache.Reset()

	// Token should still work via refresh
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Call manual Refresh
	err = provider.cache.Refresh()
	g.Expect(err).To(BeNil())

	// Token should still work
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())
}

// TestTaskStateFiltering tests that only Pending/Running tasks are cached.
func TestTaskStateFiltering(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create tasks in various states
	states := []string{"Pending", "Running", "Succeeded", "Failed", "Canceled"}
	tasks := make([]*model.Task, len(states))
	for i, state := range states {
		tasks[i] = &model.Task{
			Name:  "task-" + state,
			State: state,
		}
		err = db.Create(tasks[i]).Error
		g.Expect(err).To(BeNil())
	}

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Verify only Pending and Running are in cache
	provider.cache.mutex.RLock()
	_, foundPending := provider.cache.taskById[tasks[0].ID]
	_, foundRunning := provider.cache.taskById[tasks[1].ID]
	_, foundSucceeded := provider.cache.taskById[tasks[2].ID]
	_, foundFailed := provider.cache.taskById[tasks[3].ID]
	_, foundCanceled := provider.cache.taskById[tasks[4].ID]
	provider.cache.mutex.RUnlock()

	g.Expect(foundPending).To(BeTrue())
	g.Expect(foundRunning).To(BeTrue())
	g.Expect(foundSucceeded).To(BeFalse())
	g.Expect(foundFailed).To(BeFalse())
	g.Expect(foundCanceled).To(BeFalse())
}

// TestCacheFindSubject tests finding subjects (users and identities) by subject string.
func TestCacheFindSubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user with roles and permissions
	perm := &model.Permission{
		Name:  "Read Apps",
		Scope: "applications:GET",
	}
	err = db.Create(perm).Error
	g.Expect(err).To(BeNil())

	role := &model.Role{
		Name: "AppReader",
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	err = db.Model(role).Association("Permissions").Append(perm)
	g.Expect(err).To(BeNil())

	user := &model.User{
		Subject:  "user-subject-123",
		Userid:   "testuser",
		Password: secret.HashPassword("password"),
		Email:    "user@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	err = db.Model(user).Association("Roles").Append(role)
	g.Expect(err).To(BeNil())

	// Create test IdP identity
	identity := &Identity{
		Issuer:       "https://idp.example.com",
		Subject:      "idp-subject-456",
		RefreshToken: "refresh-token",
		Expiration:   time.Now().Add(24 * time.Hour),
		Scopes:       "openid profile email",
		Roles:        "admin developer",
		Userid:       "idpuser",
		Email:        "idp@example.com",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test finding user by subject
	subject, err := provider.cache.FindSubject(user.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeTrue())
	g.Expect(subject.IsIdentity()).To(BeFalse())
	g.Expect(subject.name).To(Equal("testuser"))
	g.Expect(subject.email).To(Equal("user@example.com"))
	g.Expect(subject.roles).To(ContainElement("AppReader"))
	g.Expect(subject.scopes).To(ContainElement("applications:GET"))

	// Test finding identity by subject
	subject, err = provider.cache.FindSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeFalse())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.name).To(Equal("idpuser"))
	g.Expect(subject.email).To(Equal("idp@example.com"))
	g.Expect(subject.roles).To(ContainElement("admin"))
	g.Expect(subject.roles).To(ContainElement("developer"))
	g.Expect(subject.scopes).To(ContainElement("openid"))
	g.Expect(subject.scopes).To(ContainElement("profile"))
	g.Expect(subject.scopes).To(ContainElement("email"))
}

// TestCacheFindSubjectNotFound tests NotFound error when subject doesn't exist.
func TestCacheFindSubjectNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Try to find non-existent subject
	_, err = provider.cache.FindSubject("non-existent-subject")
	g.Expect(err).NotTo(BeNil())

	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("subject"))
	g.Expect(notFound.Id).To(Equal("non-existent-subject"))
}

// TestCacheFindSubjectAutoRefresh tests refresh-on-miss behavior.
func TestCacheFindSubjectAutoRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create user after cache is initialized
	user := &model.User{
		Subject:  "new-user-subject",
		Userid:   "newuser",
		Password: secret.HashPassword("password"),
		Email:    "newuser@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// FindSubject should trigger refresh and find the new user
	subject, err := provider.cache.FindSubject(user.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeTrue())
	g.Expect(subject.name).To(Equal("newuser"))

	// Create identity after cache refresh
	identity := &Identity{
		Issuer:  "https://idp.example.com",
		Subject: "new-identity-subject",
		Userid:  "newidentity",
		Email:   "newidentity@example.com",
		Scopes:  "openid",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	// FindSubject should trigger refresh and find the new identity
	subject, err = provider.cache.FindSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.name).To(Equal("newidentity"))
}

// TestCacheFindSubjectTimeBasedRefresh tests time-based refresh with FindSubject.
func TestCacheFindSubjectTimeBasedRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Save original cache lifespan and restore after test
	originalLifespan := Settings.CacheLifespan
	defer func() {
		Settings.CacheLifespan = originalLifespan
	}()

	// Set very short cache lifespan
	Settings.CacheLifespan = 100 * time.Millisecond

	user := &model.User{
		Subject:  "time-subject-user",
		Userid:   "timesubjectuser",
		Password: secret.HashPassword("password"),
		Email:    "timesubject@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Find subject successfully (cache is fresh)
	subject, err := provider.cache.FindSubject(user.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.name).To(Equal("timesubjectuser"))

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Create new user while cache is stale
	newUser := &model.User{
		Subject:  "new-time-subject",
		Userid:   "newtimesubject",
		Password: secret.HashPassword("password"),
		Email:    "newtimesubject@example.com",
	}
	err = db.Create(newUser).Error
	g.Expect(err).To(BeNil())

	// FindSubject should trigger time-based refresh and find new user
	subject, err = provider.cache.FindSubject(newUser.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.name).To(Equal("newtimesubject"))
}

// TestCacheUserSavedBySubject tests that UserSaved updates bySubject map.
func TestCacheUserSavedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	// Create user not in DB (just for cache testing)
	user := &User{
		Model:    model.Model{ID: 999},
		Subject:  "cached-user-subject",
		Userid:   "cacheduser",
		Email:    "cached@example.com",
		Password: secret.HashPassword("password"),
	}

	// Save to cache
	cache.UserSaved(user)

	// Verify it's in both maps
	cache.mutex.RLock()
	userById, foundById := cache.userById[999]
	userBySubject, foundBySubject := cache.userBySubject["cached-user-subject"]
	cache.mutex.RUnlock()

	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())
	g.Expect(userById.Userid).To(Equal("cacheduser"))
	g.Expect(userBySubject.Userid).To(Equal("cacheduser"))

	// Verify FindSubject works without DB query
	subject, err := cache.FindSubject("cached-user-subject")
	g.Expect(err).To(BeNil())
	g.Expect(subject.name).To(Equal("cacheduser"))
}

// TestCacheIdentitySavedBySubject tests that IdentitySaved updates bySubject map.
func TestCacheIdentitySavedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	// Create identity not in DB (just for cache testing)
	identity := &Identity{
		Model:   model.Model{ID: 888},
		Issuer:  "https://test.idp.com",
		Subject: "cached-identity-subject",
		Userid:  "cachedidentity",
		Email:   "cachedidentity@example.com",
		Scopes:  "openid profile",
	}

	// Save to cache
	cache.IdentitySaved(identity)

	// Verify it's in both maps
	cache.mutex.RLock()
	identById, foundById := cache.identById[888]
	identBySubject, foundBySubject := cache.identBySubject["cached-identity-subject"]
	cache.mutex.RUnlock()

	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())
	g.Expect(identById.Userid).To(Equal("cachedidentity"))
	g.Expect(identBySubject.Userid).To(Equal("cachedidentity"))

	// Verify FindSubject works without DB query
	subject, err := cache.FindSubject("cached-identity-subject")
	g.Expect(err).To(BeNil())
	g.Expect(subject.name).To(Equal("cachedidentity"))
}

// TestCacheUserDeletedBySubject tests that UserDeleted removes from bySubject map.
func TestCacheUserDeletedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	user := &User{
		Model:   model.Model{ID: 777},
		Subject: "delete-user-subject",
		Userid:  "deleteuser",
	}

	cache.UserSaved(user)

	// Verify it's in both maps
	cache.mutex.RLock()
	_, foundById := cache.userById[777]
	_, foundBySubject := cache.userBySubject["delete-user-subject"]
	cache.mutex.RUnlock()
	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())

	// Delete user
	cache.UserDeleted(777)

	// Verify removed from both maps
	cache.mutex.RLock()
	_, foundById = cache.userById[777]
	_, foundBySubject = cache.userBySubject["delete-user-subject"]
	cache.mutex.RUnlock()
	g.Expect(foundById).To(BeFalse())
	g.Expect(foundBySubject).To(BeFalse())

	// Verify FindSubject returns NotFound
	_, err = cache.FindSubject("delete-user-subject")
	g.Expect(err).NotTo(BeNil())
}

// TestCacheIdentityDeletedBySubject tests that IdentityDeleted removes from bySubject map.
func TestCacheIdentityDeletedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	cache := provider.cache

	identity := &Identity{
		Model:   model.Model{ID: 666},
		Subject: "delete-identity-subject",
		Userid:  "deleteidentity",
	}

	cache.IdentitySaved(identity)

	// Verify it's in both maps
	cache.mutex.RLock()
	_, foundById := cache.identById[666]
	_, foundBySubject := cache.identBySubject["delete-identity-subject"]
	cache.mutex.RUnlock()
	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())

	// Delete identity
	cache.IdentityDeleted(666)

	// Verify removed from both maps
	cache.mutex.RLock()
	_, foundById = cache.identById[666]
	_, foundBySubject = cache.identBySubject["delete-identity-subject"]
	cache.mutex.RUnlock()
	g.Expect(foundById).To(BeFalse())
	g.Expect(foundBySubject).To(BeFalse())

	// Verify FindSubject returns NotFound
	_, err = cache.FindSubject("delete-identity-subject")
	g.Expect(err).NotTo(BeNil())
}

// TestStorageFindSubject tests Storage integration with cache FindSubject.
func TestStorageFindSubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test data
	user := &model.User{
		Subject:  "storage-user-subject",
		Userid:   "storageuser",
		Password: secret.HashPassword("password"),
		Email:    "storage@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	identity := &Identity{
		Issuer:  "https://storage.idp.com",
		Subject: "storage-identity-subject",
		Userid:  "storageidentity",
		Email:   "storageidentity@example.com",
		Scopes:  "openid",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Storage uses cache under the hood
	storage := provider.storage

	// Find user subject
	subject, err := storage.findSubject(user.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeTrue())
	g.Expect(subject.name).To(Equal("storageuser"))

	// Find identity subject
	subject, err = storage.findSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.name).To(Equal("storageidentity"))

	// Find non-existent subject
	_, err = storage.findSubject("non-existent")
	g.Expect(err).NotTo(BeNil())
}

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB() (db *gorm.DB, err error) {
	db, err = gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	if err != nil {
		return
	}

	// Auto-migrate test models
	err = db.AutoMigrate(
		&model.User{},
		&model.Task{},
		&model.Role{},
		&model.Permission{},
		&model.Token{},
		&model.Grant{},
		&model.RsaKey{},
		&Identity{},
	)
	return
}
