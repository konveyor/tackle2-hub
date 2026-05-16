package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/konveyor/tackle2-hub/internal/auth/seed"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
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

	// Notify cache about new user
	provider.Builtin.cache.UserSaved((*User)(user))

	key, err := provider.NewPAT(user.Subject, time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(key.Secret).ToNot(BeEmpty())

	task := &model.Task{
		State: "Pending",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		db.Delete(task)
	})

	// Notify cache about new task
	provider.Builtin.cache.TaskSaved((*Task)(task))

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

func _TestLDAP(t *testing.T) {
	g := NewGomegaWithT(t)

	ds := &LDAP{
		URL:      "ldap://f35a.redhat.com:389",
		BaseDN:   "dc=f35a,dc=redhat,dc=com",
		BindDN:   "uid=hub-ldap,ou=people,dc=f35a,dc=redhat,dc=com",
		Password: "hub",
	}
	ds.mapper.RuleSet = append(
		ds.mapper.RuleSet,
		MappingRule{
			And:   []string{"Engineering"},
			Roles: []string{"migrator"},
		})
	user, err := ds.Authenticate("jsmith", "dog8code")
	g.Expect(err).To(BeNil())
	g.Expect(user.Roles[0]).To(Equal("migrator"))
	t.Log(user)
}

// TestRoleMapper tests LDAP group to role mapping with realistic roles and patterns.
func TestRoleMapper(t *testing.T) {
	g := NewGomegaWithT(t)

	// Realistic role mapping configuration using both wildcards and explicit names
	mapper := &RoleMapper{
		RuleSet: []MappingRule{
			{
				Any: []string{
					"global-administrators",
					"it-admins",
					"security-admins",
					"*-admins", // Wildcard for platform-admins, konveyor-admins, tackle-admins, etc.
					"sre-team",
					"operations-team",
				},
				Roles: []string{"admin"},
			},
			{
				Any: []string{
					"*-architects", // Wildcard for engineering-architects, konveyor-architects, etc.
					"architects-council",
					"principal-engineers",
					"architecture-review-board",
				},
				Roles: []string{"architect"},
			},
			{
				And: []string{
					"konveyor-*",
					"*-migration-*",
				},
				Roles: []string{"migrator"},
			},
			{
				Any: []string{
					"migration-engineers",
					"engineering-migration-team",
					"application-migration-specialists",
				},
				Roles: []string{"migrator"},
			},
			{
				Any: []string{
					"*-managers",
					"*-leads",
					"engineering-leadership",
					"managers-all",
				},
				Roles: []string{"manager"},
			},
		},
	}

	// Test 1: Wildcard match - tackle-admins matches *-admins pattern
	groups := []string{"tackle-admins", "engineering-staff"}
	roles := mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("admin"))

	// Test 2: Explicit match - global-administrators
	groups = []string{"global-administrators", "engineering-all"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("admin"))

	// Test 3: Another wildcard - platform-admins matches *-admins
	groups = []string{"platform-admins", "engineering-platform-team"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("admin"))

	// Test 4: Explicit admin group - security-admins
	groups = []string{"security-admins", "engineering-security-team"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("admin"))

	// Test 5: Architect via wildcard - engineering-architects matches *-architects
	groups = []string{"engineering-architects", "engineering-senior-engineers"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("architect"))

	// Test 6: Architect via explicit match - principal-engineers
	groups = []string{"principal-engineers", "engineering-staff"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("architect"))

	// Test 7: Migrator via And condition - konveyor-migration-team matches both patterns
	groups = []string{"konveyor-migration-team", "engineering-all"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("migrator"))

	// Test 8: Migrator via Any (explicit) - migration-engineers
	groups = []string{"migration-engineers", "engineering-staff"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("migrator"))

	// Test 9: And condition not fully satisfied - konveyor-contributors has konveyor-* but no *-migration-*
	groups = []string{"konveyor-contributors", "engineering-staff"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(BeEmpty())

	// Test 10: Manager via wildcard - engineering-managers matches *-managers
	groups = []string{"engineering-managers", "managers-all"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("manager"))

	// Test 11: Manager via wildcard - team-leads matches *-leads
	groups = []string{"team-leads", "engineering-platform-team"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("manager"))

	// Test 12: Manager via explicit - engineering-leadership
	groups = []string{"engineering-leadership", "engineering-all"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("manager"))

	// Test 13: Multiple roles - platform-admins + engineering-architects
	groups = []string{"platform-admins", "engineering-architects", "engineering-staff"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(2))
	g.Expect(roles).To(ContainElement("admin"))
	g.Expect(roles).To(ContainElement("architect"))

	// Test 14: No special roles - regular engineer
	groups = []string{"engineering-staff", "engineering-application-team"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(BeEmpty())

	// Test 15: Empty groups list
	groups = []string{}
	roles = mapper.roles(groups)
	g.Expect(roles).To(BeEmpty())

	// Test 16: Case-sensitive wildcard - Konveyor-Admins (capital K) doesn't match *-admins
	groups = []string{"Konveyor-Admins", "engineering-staff"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(BeEmpty())

	// Test 17: Case-sensitive exact match
	groups = []string{"Global-Administrators", "engineering-staff"}
	roles = mapper.roles(groups)
	g.Expect(roles).To(BeEmpty())

	// Test 18: Empty mapping (both And and Any empty) - should skip
	emptyMapper := &RoleMapper{
		RuleSet: []MappingRule{
			{
				And:   []string{},
				Any:   []string{},
				Roles: []string{"empty-role"},
			},
		},
	}
	groups = []string{"it-admins", "engineering-staff"}
	roles = emptyMapper.roles(groups)
	g.Expect(roles).To(BeEmpty())
}

// TestRoleMapperAny tests the any() method edge cases.
func TestRoleMapperAny(t *testing.T) {
	g := NewGomegaWithT(t)

	mapper := &RoleMapper{}

	// Empty patterns - should return true (no restrictions)
	match := mapper.any([]string{}, []string{"group1", "group2"})
	g.Expect(match).To(BeTrue())

	// Empty groups - should return false
	match = mapper.any([]string{"pattern*"}, []string{})
	g.Expect(match).To(BeFalse())

	// Single pattern match
	match = mapper.any([]string{"admin*"}, []string{"admins"})
	g.Expect(match).To(BeTrue())

	// Single pattern no match
	match = mapper.any([]string{"admin*"}, []string{"users"})
	g.Expect(match).To(BeFalse())

	// Multiple patterns, first one matches
	match = mapper.any([]string{"admin*", "dev*"}, []string{"admins", "users"})
	g.Expect(match).To(BeTrue())

	// Multiple patterns, second one matches
	match = mapper.any([]string{"root*", "dev*"}, []string{"users", "developers"})
	g.Expect(match).To(BeTrue())

	// Multiple patterns, none match
	match = mapper.any([]string{"admin*", "root*"}, []string{"users", "guests"})
	g.Expect(match).To(BeFalse())

	// Pattern matches multiple groups (should still return true)
	match = mapper.any([]string{"*-team"}, []string{"dev-team", "qa-team", "ops-team"})
	g.Expect(match).To(BeTrue())
}

// TestRoleMapperAnd tests the and() method edge cases.
func TestRoleMapperAnd(t *testing.T) {
	g := NewGomegaWithT(t)

	mapper := &RoleMapper{}

	// Empty patterns - should return true (vacuous truth - no requirements to satisfy)
	match := mapper.and([]string{}, []string{"group1", "group2"})
	g.Expect(match).To(BeTrue())

	// Empty groups with patterns - should return false (requirements but nothing to match)
	match = mapper.and([]string{"pattern*"}, []string{})
	g.Expect(match).To(BeFalse())

	// Single pattern matches at least one group
	match = mapper.and([]string{"admin*"}, []string{"admins", "users"})
	g.Expect(match).To(BeTrue())

	// Single pattern matches no groups
	match = mapper.and([]string{"admin*"}, []string{"users", "guests"})
	g.Expect(match).To(BeFalse())

	// All patterns match - each pattern has at least one matching group
	match = mapper.and([]string{"dev*", "*team"}, []string{"developers", "qa-team"})
	g.Expect(match).To(BeTrue())

	// First pattern matches, second doesn't
	match = mapper.and([]string{"dev*", "*admin"}, []string{"developers", "qa-team"})
	g.Expect(match).To(BeFalse())

	// Second pattern matches, first doesn't
	match = mapper.and([]string{"admin*", "*team"}, []string{"developers", "qa-team"})
	g.Expect(match).To(BeFalse())

	// Both patterns match the same group
	match = mapper.and([]string{"dev*", "*team"}, []string{"dev-team"})
	g.Expect(match).To(BeTrue())

	// Multiple patterns, all match different groups
	match = mapper.and([]string{"*-admin", "dev-*", "qa-*"}, []string{"cluster-admin", "dev-team", "qa-lead"})
	g.Expect(match).To(BeTrue())
}

// TestRoleMapperIntegration tests realistic complex role mapping scenarios.
func TestRoleMapperIntegration(t *testing.T) {
	g := NewGomegaWithT(t)

	// Realistic enterprise role mapping configuration
	mapper := &RoleMapper{
		RuleSet: []MappingRule{
			{
				Any: []string{
					"global-administrators",
					"it-admins",
					"security-admins",
					"*-admins",
					"sre-team",
					"operations-team",
				},
				Roles: []string{"admin"},
			},
			{
				Any: []string{
					"*-architects",
					"architects-council",
					"principal-engineers",
					"architecture-review-board",
				},
				Roles: []string{"architect"},
			},
			{
				And: []string{
					"konveyor-*",
					"*-migration-*",
				},
				Roles: []string{"migrator"},
			},
			{
				Any: []string{
					"migration-engineers",
					"engineering-migration-team",
					"application-migration-specialists",
				},
				Roles: []string{"migrator"},
			},
			{
				Any: []string{
					"*-managers",
					"*-leads",
					"engineering-leadership",
					"managers-all",
				},
				Roles: []string{"manager"},
			},
		},
	}

	// Scenario 1: Senior engineering leader - has admin, architect, and manager roles
	groups := []string{
		"global-administrators",
		"engineering-architects",
		"engineering-leadership",
		"engineering-all",
	}
	roles := mapper.roles(groups)
	g.Expect(roles).To(HaveLen(3))
	g.Expect(roles).To(ContainElement("admin"))
	g.Expect(roles).To(ContainElement("architect"))
	g.Expect(roles).To(ContainElement("manager"))

	// Scenario 2: Konveyor migration specialist - migrator role via And condition
	groups = []string{
		"konveyor-migration-team",
		"engineering-staff",
		"engineering-all",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("migrator"))

	// Scenario 3: Engineering tech lead - manager role only
	groups = []string{
		"engineering-tech-leads",
		"engineering-platform-team",
		"engineering-all",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(1))
	g.Expect(roles[0]).To(Equal("manager"))

	// Scenario 4: Regular engineer - no special roles
	groups = []string{
		"engineering-staff",
		"engineering-application-team",
		"engineering-all",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(BeEmpty())

	// Scenario 5: IT admin who is also on the architecture review board
	groups = []string{
		"it-admins",
		"architecture-review-board",
		"it-operations",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(2))
	g.Expect(roles).To(ContainElement("admin"))
	g.Expect(roles).To(ContainElement("architect"))

	// Scenario 6: Product manager who is also a migration engineer
	groups = []string{
		"product-managers",
		"migration-engineers",
		"engineering-all",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(2))
	g.Expect(roles).To(ContainElement("manager"))
	g.Expect(roles).To(ContainElement("migrator"))

	// Scenario 7: SRE team member (admin) who leads a team (manager)
	groups = []string{
		"sre-team",
		"team-leads",
		"engineering-platform-team",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(2))
	g.Expect(roles).To(ContainElement("admin"))
	g.Expect(roles).To(ContainElement("manager"))

	// Scenario 8: Tackle admin with architect role
	groups = []string{
		"tackle-admins",
		"konveyor-architects",
		"tackle-team",
	}
	roles = mapper.roles(groups)
	g.Expect(roles).To(HaveLen(2))
	g.Expect(roles).To(ContainElement("admin"))
	g.Expect(roles).To(ContainElement("architect"))
}

// TestCacheAutoRefresh tests automatic cache refresh on miss and expiry.
func TestCacheNotificationPropagation(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create user BEFORE provider initialization
	user := &model.User{
		Subject:  "cache-notification-user",
		Userid:   "cachenotifuser",
		Password: secret.HashPassword("password"),
		Email:    "cachenotif@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// User should be in cache (loaded during provider initialization)
	token, err := provider.NewPAT(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Verify token works (NewPAT calls TokenSaved notification)
	request := &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Delete token from DB and notify cache
	err = db.Delete(&model.Token{}, token.ID).Error
	g.Expect(err).To(BeNil())
	provider.cache.TokenDeleted(token.ID)

	// Verify token is immediately gone from cache (notification propagated)
	request = &Request{}
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil()) // Should fail immediately, not wait for refresh
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
	err = secret.Encrypt(identity)
	g.Expect(err).To(BeNil())
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

// TestCacheTransaction tests cache transaction behavior.
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
	unboundToken := &model.Token{
		Kind:          KindAPIKey,
		Digest:        secret.Hash("unbound-token"),
		UserID:        nil,
		TaskID:        nil,
		IdpIdentityID: nil,
		Expiration:    time.Now().Add(24 * time.Hour),
	}
	err = db.Create(unboundToken).Error
	g.Expect(err).To(BeNil())

	// Add to cache
	cache.TokenSaved(&Token{Token: *unboundToken, Secret: "unbound-token"})

	m, err := cache.FindToken("unbound-token")
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
	_, errPending := provider.cache.FindTaskById(tasks[0].ID)
	_, errRunning := provider.cache.FindTaskById(tasks[1].ID)
	_, errSucceeded := provider.cache.FindTaskById(tasks[2].ID)
	_, errFailed := provider.cache.FindTaskById(tasks[3].ID)
	_, errCanceled := provider.cache.FindTaskById(tasks[4].ID)

	g.Expect(errPending).To(BeNil())
	g.Expect(errRunning).To(BeNil())
	g.Expect(errSucceeded).NotTo(BeNil())
	g.Expect(errFailed).NotTo(BeNil())
	g.Expect(errCanceled).NotTo(BeNil())
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
		Userid:       "idpuser",
		Email:        "idp@example.com",
	}
	err = secret.Encrypt(identity)
	g.Expect(err).To(BeNil())
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
	g.Expect(subject.Key).To(Equal("user-subject-123"))
	g.Expect(subject.User.Userid).To(Equal("testuser"))
	g.Expect(subject.Email).To(Equal("user@example.com"))
	g.Expect(subject.Scopes).To(ContainElement("applications:GET"))

	// Test finding identity by subject
	subject, err = provider.cache.FindSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeFalse())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.Key).To(Equal("idp-subject-456"))
	g.Expect(subject.Identity.Userid).To(Equal("idpuser"))
	g.Expect(subject.Email).To(Equal("idp@example.com"))
	g.Expect(subject.Scopes).To(ContainElement("openid"))
	g.Expect(subject.Scopes).To(ContainElement("profile"))
	g.Expect(subject.Scopes).To(ContainElement("email"))
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

// TestCacheFindSubjectMiss tests lazy-load and notification behavior.
func TestCacheFindSubjectMiss(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Force initial cache load by accessing a non-existent subject
	_, _ = provider.cache.FindSubject("force-initial-load")

	// Create user after cache is loaded (NOT notified to cache)
	user := &model.User{
		Subject:  "new-user-subject",
		Userid:   "newuser",
		Password: secret.HashPassword("password"),
		Email:    "newuser@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// FindSubject should NOT find it (cache is fresh, no notification)
	subject, err := provider.cache.FindSubject(user.Subject)
	g.Expect(err).NotTo(BeNil()) // NotFound
	g.Expect(subject).To(BeNil())

	// But if we notify the cache, it should be found immediately
	provider.cache.UserSaved((*User)(user))
	subject, err = provider.cache.FindSubject(user.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeTrue())
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
	g.Expect(subject.Key).To(Equal("time-subject-user"))
	g.Expect(subject.User.Userid).To(Equal("timesubjectuser"))

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
	g.Expect(subject.Key).To(Equal("new-time-subject"))
	g.Expect(subject.User.Userid).To(Equal("newtimesubject"))
}

// TestCacheUserSavedBySubject tests that UserSaved updates bySubject map.
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
	err = secret.Encrypt(identity)
	g.Expect(err).To(BeNil())
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
	g.Expect(subject.Key).To(Equal("storage-user-subject"))
	g.Expect(subject.User.Userid).To(Equal("storageuser"))

	// Find identity subject
	subject, err = storage.findSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.Key).To(Equal("storage-identity-subject"))
	g.Expect(subject.Identity.Userid).To(Equal("storageidentity"))

	// Find non-existent subject
	_, err = storage.findSubject("non-existent")
	g.Expect(err).NotTo(BeNil())
}

// TestCacheFindUserByUserid tests finding user by userid field.
func TestCacheFindUserByUserid(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user
	user := &model.User{
		Subject:  "userid-test-subject",
		Userid:   "testuserid",
		Password: secret.HashPassword("password"),
		Email:    "userid@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Find by userid
	found, err := provider.cache.FindUserByUserid("testuserid")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(Equal("userid-test-subject"))
	g.Expect(found.Email).To(Equal("userid@example.com"))

	// Find non-existent userid
	_, err = provider.cache.FindUserByUserid("nonexistent")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("user"))
}

// TestCacheFindUserByUseridNotification tests notification-based cache updates.
func TestCacheFindUserByUseridNotification(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Force initial cache load
	_, _ = provider.cache.FindUserByUserid("force-initial-load")

	// Create user after cache is loaded (NOT notified)
	user := &model.User{
		Subject:  "new-userid-subject",
		Userid:   "newuserid",
		Password: secret.HashPassword("password"),
		Email:    "newuserid@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// FindUserByUserid should NOT find it (cache is fresh, no notification)
	found, err := provider.cache.FindUserByUserid("newuserid")
	g.Expect(err).NotTo(BeNil()) // NotFound
	g.Expect(found).To(BeNil())

	// Notify cache of user creation
	provider.cache.UserSaved((*User)(user))

	// Now it should be found immediately
	found, err = provider.cache.FindUserByUserid("newuserid")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(Equal("new-userid-subject"))
}

// TestCacheFindUserByUseridTimeRefresh tests time-based refresh for userid lookup.
func TestCacheFindUserByUseridTimeRefresh(t *testing.T) {
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
		Subject:  "time-userid-subject",
		Userid:   "timeuserid",
		Password: secret.HashPassword("password"),
		Email:    "timeuserid@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Find successfully (cache is fresh)
	found, err := provider.cache.FindUserByUserid("timeuserid")
	g.Expect(err).To(BeNil())
	g.Expect(found.Subject).To(Equal("time-userid-subject"))

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Create new user while cache is stale
	newUser := &model.User{
		Subject:  "new-time-userid-subject",
		Userid:   "newtimeuserid",
		Password: secret.HashPassword("password"),
		Email:    "newtimeuserid@example.com",
	}
	err = db.Create(newUser).Error
	g.Expect(err).To(BeNil())

	// FindUserByUserid should trigger time-based refresh
	found, err = provider.cache.FindUserByUserid("newtimeuserid")
	g.Expect(err).To(BeNil())
	g.Expect(found.Subject).To(Equal("new-time-userid-subject"))
}

// TestCacheGetTask tests finding task by ID.
func TestCacheGetTask(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task
	task := &model.Task{
		Name:  "cache-test-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Get task by ID
	found, err := provider.cache.FindTaskById(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Name).To(Equal("cache-test-task"))
	g.Expect(found.State).To(Equal("Running"))

	// Get non-existent task
	_, err = provider.cache.FindTaskById(9999)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("task"))
}

// TestCacheFindTaskByIdNotification tests notification-based cache updates for tasks.
func TestCacheFindTaskByIdNotification(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Force initial cache load
	_, _ = provider.cache.FindTaskById(9999)

	// Create task after cache is loaded (NOT notified)
	task := &model.Task{
		Name:  "new-cache-task",
		State: "Pending",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// FindTaskById should NOT find it (cache is fresh, no notification)
	found, err := provider.cache.FindTaskById(task.ID)
	g.Expect(err).NotTo(BeNil()) // NotFound
	g.Expect(found).To(BeNil())

	// Notify cache of task creation
	provider.cache.TaskSaved((*Task)(task))

	// Now it should be found immediately
	found, err = provider.cache.FindTaskById(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Name).To(Equal("new-cache-task"))
}

// TestCacheGetTaskTimeRefresh tests time-based refresh for task lookup.
func TestCacheGetTaskTimeRefresh(t *testing.T) {
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

	task := &model.Task{
		Name:  "time-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Get successfully (cache is fresh)
	found, err := provider.cache.FindTaskById(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(found.Name).To(Equal("time-task"))

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Create new task while cache is stale
	newTask := &model.Task{
		Name:  "new-time-task",
		State: "Pending",
	}
	err = db.Create(newTask).Error
	g.Expect(err).To(BeNil())

	// GetTask should trigger time-based refresh
	found, err = provider.cache.FindTaskById(newTask.ID)
	g.Expect(err).To(BeNil())
	g.Expect(found.Name).To(Equal("new-time-task"))
}

// TestCacheUserByUseridMaps tests that all userid maps are maintained.
func TestStoreDeviceAuthorization(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	storage := provider.storage

	clientId := "test-client"
	deviceCode := "device-code-123"
	userCode := "ABCD-1234"
	expires := time.Now().Add(15 * time.Minute)
	scopes := []string{"openid", "profile"}

	err = storage.StoreDeviceAuthorization(
		context.Background(),
		clientId,
		deviceCode,
		userCode,
		expires,
		scopes,
	)
	g.Expect(err).To(BeNil())

	// Verify device authorization was created in memory
	devAuth, found := storage.GetDevAuthByUserCode(userCode)
	g.Expect(found).To(BeTrue())
	g.Expect(devAuth.clientId).To(Equal(clientId))
	g.Expect(devAuth.userCode).To(Equal(userCode))
	g.Expect(devAuth.deviceCode).To(Equal(deviceCode))
	g.Expect(devAuth.scopes).To(Equal(scopes))
	g.Expect(devAuth.done).To(BeFalse())
	g.Expect(devAuth.denied).To(BeFalse())
}

// TestGetDeviceAuthorizatonStatePending tests retrieving pending device grant.
func TestGetDeviceAuthorizatonStatePending(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	storage := provider.storage

	clientId := "test-client"
	deviceCode := "device-code-pending"
	userCode := "WXYZ-5678"
	expires := time.Now().Add(15 * time.Minute)
	scopes := []string{"openid"}

	err = storage.StoreDeviceAuthorization(
		context.Background(),
		clientId,
		deviceCode,
		userCode,
		expires,
		scopes,
	)
	g.Expect(err).To(BeNil())

	state, err := storage.GetDeviceAuthorizatonState(
		context.Background(),
		clientId,
		deviceCode,
	)
	g.Expect(err).To(BeNil())
	g.Expect(state.ClientID).To(Equal(clientId))
	g.Expect(state.Scopes).To(Equal([]string{"openid"}))
	g.Expect(state.Done).To(BeFalse())
	g.Expect(state.Denied).To(BeFalse())
	g.Expect(state.Subject).To(Equal(""))
}

// TestGetDeviceAuthorizatonStateDone tests retrieving authorized device grant.
func TestGetDeviceAuthorizatonStateDone(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	storage := provider.storage

	clientId := "test-client"
	deviceCode := "device-code-done"
	userCode := "DONE-1234"
	expires := time.Now().Add(15 * time.Minute)
	scopes := []string{"openid", "profile"}

	err = storage.StoreDeviceAuthorization(
		context.Background(),
		clientId,
		deviceCode,
		userCode,
		expires,
		scopes,
	)
	g.Expect(err).To(BeNil())

	// Authorize the device
	authTime := time.Now().Truncate(time.Second)
	err = storage.UpdateDevAuth(userCode, "authorized-user", true, false, authTime)
	g.Expect(err).To(BeNil())

	state, err := storage.GetDeviceAuthorizatonState(
		context.Background(),
		clientId,
		deviceCode,
	)
	g.Expect(err).To(BeNil())
	g.Expect(state.Done).To(BeTrue())
	g.Expect(state.Denied).To(BeFalse())
	g.Expect(state.Subject).To(Equal("authorized-user"))
	g.Expect(state.AuthTime.Unix()).To(Equal(authTime.Unix()))
}

// TestGetDeviceAuthorizatonStateDenied tests retrieving denied device grant.
func TestGetDeviceAuthorizatonStateDenied(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	storage := provider.storage

	clientId := "test-client"
	deviceCode := "device-code-denied"
	userCode := "DENY-9999"
	expires := time.Now().Add(15 * time.Minute)
	scopes := []string{"openid"}

	err = storage.StoreDeviceAuthorization(
		context.Background(),
		clientId,
		deviceCode,
		userCode,
		expires,
		scopes,
	)
	g.Expect(err).To(BeNil())

	// Deny the device authorization
	err = storage.UpdateDevAuth(userCode, "", false, true, time.Time{})
	g.Expect(err).To(BeNil())

	state, err := storage.GetDeviceAuthorizatonState(
		context.Background(),
		clientId,
		deviceCode,
	)
	g.Expect(err).To(BeNil())
	g.Expect(state.Done).To(BeFalse())
	g.Expect(state.Denied).To(BeTrue())
}

// TestGetDeviceAuthorizatonStateNotFound tests invalid device code.
func TestGetDeviceAuthorizatonStateNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())
	storage := provider.storage

	_, err = storage.GetDeviceAuthorizatonState(
		context.Background(),
		"test-client",
		"invalid-device-code",
	)
	g.Expect(err).NotTo(BeNil())
}

// TestReadClients tests reading clients from clients.yaml.
func TestReadClients(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	domain := NewDomain(db)

	clients, err := domain.readClients()
	g.Expect(err).To(BeNil())
	g.Expect(clients).To(HaveLen(3))

	// Verify web-ui client
	webUI := findClientByID(clients, "web-ui")
	g.Expect(webUI).NotTo(BeNil())
	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(webUI.ApplicationType).To(Equal("web"))
	g.Expect(webUI.Grants).To(ContainElement("authorization_code"))
	g.Expect(webUI.Scopes).To(ContainElement("openid"))

	// Verify kantra client
	kantra := findClientByID(clients, "kantra")
	g.Expect(kantra).NotTo(BeNil())
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kantra.ApplicationType).To(Equal("native"))
	g.Expect(kantra.Grants).To(ContainElement("urn:ietf:params:oauth:grant-type:device_code"))

	// Verify kai-ide client
	kaiIDE := findClientByID(clients, "kai-ide")
	g.Expect(kaiIDE).NotTo(BeNil())
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))
	g.Expect(kaiIDE.ApplicationType).To(Equal("native"))
}

// TestSeedClientsCreate tests creating new clients from YAML.
func TestSeedClientsCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	// Set redirect URIs in settings
	originalWebUI := Settings.Auth.RedirectURI.WebUI
	originalKAI := Settings.Auth.RedirectURI.KAI
	defer func() {
		Settings.Auth.RedirectURI.WebUI = originalWebUI
		Settings.Auth.RedirectURI.KAI = originalKAI
	}()

	Settings.Auth.RedirectURI.WebUI = "http://localhost:3000/login/callback"
	Settings.Auth.RedirectURI.KAI = "vscode://test.extension/auth"

	domain := NewDomain(db)

	// Seed clients
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify clients were created in database
	var clients []IdpClient
	err = db.Find(&clients).Error
	g.Expect(err).To(BeNil())
	g.Expect(clients).To(HaveLen(3))

	// Verify web-ui client with injected redirect URI
	var webUI IdpClient
	err = db.First(&webUI, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(webUI.RedirectURIs).To(HaveLen(1))
	g.Expect(webUI.RedirectURIs[0]).To(Equal("http://localhost:3000/login/callback"))

	// Verify kantra client (no redirect URIs)
	var kantra IdpClient
	err = db.First(&kantra, "ClientId = ?", "kantra").Error
	g.Expect(err).To(BeNil())
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kantra.RedirectURIs).To(BeEmpty())

	// Verify kai-ide client with injected redirect URI
	var kaiIDE IdpClient
	err = db.First(&kaiIDE, "ClientId = ?", "kai-ide").Error
	g.Expect(err).To(BeNil())
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))
	g.Expect(kaiIDE.RedirectURIs).To(HaveLen(1))
	g.Expect(kaiIDE.RedirectURIs[0]).To(Equal("vscode://test.extension/auth"))
}

// TestSeedClientsUpdate tests updating existing clients.
func TestSeedClientsUpdate(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	// Create existing client with different grants
	existing := &IdpClient{
		ClientId:        "web-ui",
		ApplicationType: "web",
		Grants:          []string{"authorization_code"}, // Missing some grants
		Scopes:          []string{"openid"},
	}
	existing.ID = 1
	err = db.Create(existing).Error
	g.Expect(err).To(BeNil())

	// Set redirect URI
	originalWebUI := Settings.Auth.RedirectURI.WebUI
	defer func() {
		Settings.Auth.RedirectURI.WebUI = originalWebUI
	}()
	Settings.Auth.RedirectURI.WebUI = "http://localhost:3000/login/callback"

	domain := NewDomain(db)

	// Seed clients (should update existing)
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify client was updated
	var updated IdpClient
	err = db.First(&updated, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	g.Expect(updated.ID).To(Equal(uint(1))) // ID preserved
	g.Expect(updated.Grants).To(HaveLen(3)) // Updated from YAML
	g.Expect(updated.Grants).To(ContainElement("urn:ietf:params:oauth:grant-type:jwt-bearer"))
	g.Expect(updated.RedirectURIs).To(HaveLen(1))
	g.Expect(updated.RedirectURIs[0]).To(Equal("http://localhost:3000/login/callback"))
}

// TestSeedClientsDelete tests deleting orphaned seeded clients.
func TestSeedClientsDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	// Create orphaned seeded client (ID < 1000, not in YAML)
	orphaned := &IdpClient{
		ClientId:        "orphaned-client",
		ApplicationType: "web",
		Grants:          []string{"authorization_code"},
		Scopes:          []string{"openid"},
	}
	orphaned.ID = 999 // ID < LastId
	err = db.Create(orphaned).Error
	g.Expect(err).To(BeNil())

	// Create non-seeded client (ID >= 1000, should be preserved)
	nonSeeded := &IdpClient{
		ClientId:        "custom-client",
		ApplicationType: "native",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"api"},
	}
	nonSeeded.ID = 1001 // ID >= LastId
	err = db.Create(nonSeeded).Error
	g.Expect(err).To(BeNil())

	domain := NewDomain(db)

	// Seed clients
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify orphaned client was deleted
	var orphanedCheck IdpClient
	err = db.First(&orphanedCheck, "ClientId = ?", "orphaned-client").Error
	g.Expect(err).NotTo(BeNil()) // Should not be found

	// Verify non-seeded client was preserved
	var nonSeededCheck IdpClient
	err = db.First(&nonSeededCheck, "ClientId = ?", "custom-client").Error
	g.Expect(err).To(BeNil())
	g.Expect(nonSeededCheck.ID).To(Equal(uint(1001)))

	// Verify YAML clients were created
	var count int64
	db.Model(&IdpClient{}).Count(&count)
	g.Expect(count).To(Equal(int64(4))) // 3 from YAML + 1 non-seeded
}

// TestClientPatch tests the client reconciliation patch logic.
func TestClientPatch(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	domain := NewDomain(db)

	// Create existing clients
	webUI := IdpClient{
		ClientId:        "web-ui",
		ApplicationType: "web",
		Grants:          []string{"old-grant"},
	}
	webUI.ID = 1
	orphaned := IdpClient{
		ClientId:        "orphaned",
		ApplicationType: "native",
	}
	orphaned.ID = 999
	existing := map[string]IdpClient{
		"web-ui":   webUI,
		"orphaned": orphaned,
	}

	// Wanted clients from YAML
	wanted := []seed.IdpClient{
		{
			ID:              1,
			ClientId:        "web-ui",
			ApplicationType: "web",
			Grants:          []string{"new-grant"},
			RedirectURIs:    []string{"http://localhost/callback"},
		},
		{
			ID:              2,
			ClientId:        "kantra",
			ApplicationType: "native",
			Grants:          []string{"device_code"},
		},
	}

	patch := domain.clientPatch(existing, wanted)

	// Verify patch operations
	g.Expect(patch.toDelete).To(HaveLen(1))
	g.Expect(patch.toDelete[0]).To(Equal(uint(999))) // orphaned

	g.Expect(patch.toUpdate).To(HaveLen(1))
	g.Expect(patch.toUpdate[0].client.ClientId).To(Equal("web-ui"))
	g.Expect(patch.toUpdate[0].seed.Grants).To(Equal([]string{"new-grant"}))

	g.Expect(patch.toCreate).To(HaveLen(1))
	g.Expect(patch.toCreate[0].client.ClientId).To(Equal("kantra"))
	g.Expect(patch.toCreate[0].client.ID).To(Equal(uint(2))) // ID from YAML
}

// TestSeedClientsRedirectURIInjection tests redirect URI injection.
func TestSeedClientsRedirectURIInjection(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	// Test with custom redirect URIs
	originalWebUI := Settings.Auth.RedirectURI.WebUI
	originalKAI := Settings.Auth.RedirectURI.KAI
	defer func() {
		Settings.Auth.RedirectURI.WebUI = originalWebUI
		Settings.Auth.RedirectURI.KAI = originalKAI
	}()

	Settings.Auth.RedirectURI.WebUI = "https://custom.example.com/callback"
	Settings.Auth.RedirectURI.KAI = "vscode://custom.extension/auth"

	domain := NewDomain(db)
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify custom redirect URIs were injected
	var webUI IdpClient
	err = db.First(&webUI, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	g.Expect(webUI.RedirectURIs).To(Equal([]string{"https://custom.example.com/callback"}))

	var kaiIDE IdpClient
	err = db.First(&kaiIDE, "ClientId = ?", "kai-ide").Error
	g.Expect(err).To(BeNil())
	g.Expect(kaiIDE.RedirectURIs).To(Equal([]string{"vscode://custom.extension/auth"}))

	// Verify kantra has no redirect URIs
	var kantra IdpClient
	err = db.First(&kantra, "ClientId = ?", "kantra").Error
	g.Expect(err).To(BeNil())
	g.Expect(kantra.RedirectURIs).To(BeEmpty())
}

// TestSeedClientsIDPreservation tests that IDs from YAML are preserved.
func TestSeedClientsIDPreservation(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupDomainTestDB()
	g.Expect(err).To(BeNil())

	domain := NewDomain(db)

	// First seed
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify IDs
	var webUI, kantra, kaiIDE IdpClient
	db.First(&webUI, "ClientId = ?", "web-ui")
	db.First(&kantra, "ClientId = ?", "kantra")
	db.First(&kaiIDE, "ClientId = ?", "kai-ide")

	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))

	// Second seed (should preserve IDs)
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	db.First(&webUI, "ClientId = ?", "web-ui")
	db.First(&kantra, "ClientId = ?", "kantra")
	db.First(&kaiIDE, "ClientId = ?", "kai-ide")

	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))
}

// setupDomainTestDB creates an in-memory SQLite database for domain testing.
func setupDomainTestDB() (db *gorm.DB, err error) {
	db, err = database.OpenTest()
	if err != nil {
		return
	}

	// Auto-migrate IdpClient model
	err = db.AutoMigrate(&model.IdpClient{})
	return
}

// findClientByID finds a client in a slice by ClientId.
func findClientByID(clients []seed.IdpClient, clientId string) *seed.IdpClient {
	for i := range clients {
		if clients[i].ClientId == clientId {
			return &clients[i]
		}
	}
	return nil
}

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB() (db *gorm.DB, err error) {
	db, err = database.OpenTest()
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
