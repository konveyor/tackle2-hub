package auth

import (
	"errors"
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

// TestUserKey tests creating and authenticating with user API keys.
func TestUserKey(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user
	user := &model.User{
		UUID:     "test-uuid-123",
		UserId:   "testuser",
		Password: "testpassword",
		Email:    "test@example.com",
	}
	err = secret.Encrypt(user)
	g.Expect(err).To(BeNil())
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test creating API key with valid credentials
	kr := KeyRequest{
		Userid:   user.UserId,
		Password: "testpassword",
		Lifespan: 24 * time.Hour,
	}
	key, err := provider.Grant(kr)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).NotTo(BeEmpty())
	g.Expect(key.Digest).NotTo(BeEmpty())
	g.Expect(key.Digest).To(Equal(hashSecret(key.Secret)))

	// Verify the API key was created in the database
	var keyCount int64
	err = db.Model(&model.APIKey{}).Count(&keyCount).Error
	g.Expect(err).To(BeNil())
	g.Expect(keyCount).To(Equal(int64(1)))

	// Verify the digest in the database matches
	var dbKey model.APIKey
	err = db.First(&dbKey).Error
	g.Expect(err).To(BeNil())
	g.Expect(dbKey.Digest).To(Equal(key.Digest))

	// Test creating API key with invalid password
	kr.Password = "wrong-password"
	_, err = provider.Grant(kr)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test creating API key with non-existent user
	kr.Userid = "not-existing-user"
	_, err = provider.Grant(kr)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test authenticating with the API key
	request := &Request{
		Token: "Bearer " + key.Secret,
	}
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify token claims
	claims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(claims[ClaimSub]).To(Equal("test-uuid-123"))

	// Test that expired keys are rejected
	expiredSecret := "expired-secret-key"
	expiredKey := &model.APIKey{
		UserID:     &user.ID,
		Expiration: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Digest:     hashSecret(expiredSecret),
	}
	err = db.Create(expiredKey).Error
	g.Expect(err).To(BeNil())

	request = &Request{
		Token: "Bearer " + expiredSecret,
	}
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestTaskKey tests creating and authenticating with task API keys.
func TestTaskKey(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task
	task := &model.Task{
		Name: "test-task",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test creating task API key
	kr := KeyRequest{
		TaskID:   task.ID,
		Lifespan: 24 * time.Hour,
	}
	key, err := provider.Grant(kr)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).NotTo(BeEmpty())
	g.Expect(key.Digest).NotTo(BeEmpty())
	g.Expect(key.Digest).To(Equal(hashSecret(key.Secret)))

	// Test authenticating with the task API key
	request := &Request{
		Token: "Bearer " + key.Secret,
	}
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify the digest in the database matches
	var dbKey model.APIKey
	err = db.First(&dbKey).Error
	g.Expect(err).To(BeNil())
	g.Expect(dbKey.Digest).To(Equal(key.Digest))

	// Test creating key for non-existent task
	kr.TaskID = 9999
	_, err = provider.Grant(kr)
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
	request := &Request{
		Token: "Bearer " + tokenString,
	}
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

	request = &Request{
		Token: "Bearer " + expiredTokenString,
	}
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

	request = &Request{
		Token: "Bearer " + noSubTokenString,
	}
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

	request = &Request{
		Token: "Bearer " + noScopeTokenString,
	}
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
	request := &Request{
		Token: "invalid-token-without-bearer",
	}
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test empty token
	request = &Request{
		Token: "",
	}
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test malformed bearer token
	request = &Request{
		Token: "Bearer",
	}
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test invalid JWT
	request = &Request{
		Token: "Bearer invalid.jwt.token",
	}
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

// TestHashSecretDeterministic tests that hashing is deterministic.
func TestHashSecretDeterministic(t *testing.T) {
	g := NewGomegaWithT(t)

	secret := "test-secret-key"

	// Same input should produce same hash
	hash1 := hashSecret(secret)
	hash2 := hashSecret(secret)
	g.Expect(hash1).To(Equal(hash2))

	// Different inputs should produce different hashes
	hash3 := hashSecret("different-secret")
	g.Expect(hash1).NotTo(Equal(hash3))

	// Hash should not be empty
	g.Expect(hash1).NotTo(BeEmpty())
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

	// Create API key for running task - should work
	kr := KeyRequest{
		TaskID:   task.ID,
		Lifespan: 24 * time.Hour,
	}
	key, err := provider.Grant(kr)
	g.Expect(err).To(BeNil())

	// Authenticate with key - should work
	request := &Request{Token: "Bearer " + key.Secret}
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Update task to Succeeded - key should now be rejected
	db.Model(task).Update("State", "Succeeded")
	request = &Request{Token: "Bearer " + key.Secret}
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test with Failed state
	db.Model(task).Update("State", "Failed")
	request = &Request{Token: "Bearer " + key.Secret}
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test with Canceled state
	db.Model(task).Update("State", "Canceled")
	request = &Request{Token: "Bearer " + key.Secret}
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
		UUID:     "user-123",
		UserId:   "testuser",
		Password: "password",
		Email:    "test@example.com",
	}
	err = secret.Encrypt(user)
	g.Expect(err).To(BeNil())
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
	Hub = provider

	// Create API key
	kr := KeyRequest{
		Userid:   "testuser",
		Password: "password",
		Lifespan: 24 * time.Hour,
	}
	key, err := provider.Grant(kr)
	g.Expect(err).To(BeNil())

	// Test authenticated and authorized (matching scope)
	request := &Request{
		Token:  "Bearer " + key.Secret,
		Scope:  "applications",
		Method: "GET",
		DB:     db,
	}
	result, err := request.Permit()
	g.Expect(err).To(BeNil())
	g.Expect(result.Authenticated).To(BeTrue())
	g.Expect(result.Authorized).To(BeTrue())
	g.Expect(result.User).To(Equal("user-123"))

	// Test authenticated but not authorized (wrong method)
	request = &Request{
		Token:  "Bearer " + key.Secret,
		Scope:  "applications",
		Method: "POST",
		DB:     db,
	}
	result, err = request.Permit()
	g.Expect(err).To(BeNil())
	g.Expect(result.Authenticated).To(BeTrue())
	g.Expect(result.Authorized).To(BeFalse())

	// Test not authenticated (invalid token)
	request = &Request{
		Token:  "Bearer invalid-token",
		Scope:  "applications",
		Method: "GET",
		DB:     db,
	}
	result, err = request.Permit()
	g.Expect(err).To(BeNil())
	g.Expect(result.Authenticated).To(BeFalse())
	g.Expect(result.Authorized).To(BeFalse())
}

// TestNoAuthProvider tests the NoAuth provider fallback behavior.
func TestNoAuthProvider(t *testing.T) {
	g := NewGomegaWithT(t)

	provider := &NoAuth{}

	// Authenticate always succeeds (returns nil token, nil error)
	request := &Request{Token: "any-token"}
	token, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Scopes returns wildcard
	scopes := provider.Scopes(token)
	g.Expect(scopes).To(HaveLen(1))
	scope := scopes[0]
	g.Expect(scope.Match("anything", "GET")).To(BeTrue())
	g.Expect(scope.Match("anything", "POST")).To(BeTrue())

	// User returns fixed admin user
	user := provider.User(token)
	g.Expect(user).To(Equal("admin.noauth"))

	// UserKey and TaskKey return empty (no-op)
	kr := KeyRequest{
		Userid:   "user-123",
		Password: "password-123",
		Lifespan: time.Hour,
	}
	key, err := provider.Grant(kr)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).To(BeEmpty())

	kr.TaskID = 1
	key, err = provider.Grant(kr)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).To(BeEmpty())
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
		&model.APIKey{},
		&model.Grant{},
		&model.RsaKey{},
	)
	return
}
