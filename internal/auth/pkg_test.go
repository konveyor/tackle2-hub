package auth

import (
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
	key, err := provider.UserKey("testuser", "testpassword", 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).NotTo(BeEmpty())

	// Verify the API key was created in the database
	var keyCount int64
	err = db.Model(&model.APIKey{}).Count(&keyCount).Error
	g.Expect(err).To(BeNil())
	g.Expect(keyCount).To(Equal(int64(1)))

	// Test creating API key with invalid password
	_, err = provider.UserKey("testuser", "wrongpassword", 24*time.Hour)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test creating API key with non-existent user
	_, err = provider.UserKey("nonexistent", "password", 24*time.Hour)
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
	expiredKey := &model.APIKey{
		UserID:     &user.ID,
		Expiration: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Digest:     "expired-secret-key",
	}
	err = secret.Encrypt(expiredKey)
	g.Expect(err).To(BeNil())
	err = db.Create(expiredKey).Error
	g.Expect(err).To(BeNil())

	request = &Request{
		Token: "Bearer expired-secret-key",
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
	key, err := provider.TaskKey(task.ID, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).NotTo(BeEmpty())

	// Test authenticating with the task API key
	request := &Request{
		Token: "Bearer " + key.Secret,
	}
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Test creating key for non-existent task
	_, err = provider.TaskKey(99999, 24*time.Hour)
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
