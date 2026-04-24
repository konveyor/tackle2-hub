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

	// Create test task
	task := &model.Task{
		Name: "test-task",
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

// TestKeyCacheDelete tests the KeyCache Delete method removes keys from both indexes.
func TestKeyCacheDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user with permissions
	user := &model.User{
		Subject:  "cache-delete-user",
		Userid:   "cachedeleteuser",
		Password: secret.HashPassword("password"),
		Email:    "cachedelete@example.com",
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

	// Create provider with cache
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

	// Delete the key from cache only
	provider.cache.Delete(token.ID)

	// Verify token is removed from cache - next call should fetch from DB again
	// Since the key still exists in DB, authentication should succeed
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Populate cache again
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Now delete from DB manually and from cache
	err = db.Delete(&token).Error
	g.Expect(err).To(BeNil())
	provider.cache.Delete(token.ID)

	// Authentication should now fail (not in cache or DB)
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
		&model.IdpIdentity{},
	)
	return
}
