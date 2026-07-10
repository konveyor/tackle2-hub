package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	. "github.com/onsi/gomega"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// testIssuer returns the issuer URL for test requests.
func testIssuer() string {
	return "http://localhost:8080/oidc"
}

// newTestRequest creates a Request with a gin.Context containing an HTTP request.
// This provides the request context needed for issuer validation in token operations.
func newTestRequest() (req *Request) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "http://localhost:8080/test", nil)

	req = &Request{}
	req.CTX = ctx
	return
}

// TestUserGrant tests creating and authenticating with user tokens.
func TestUserGrant(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user
	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "testuser",
		Password: secret.HashPassword("testpassword"),
		Email:    "test@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get auto-generated subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())
	g.Expect(user.Subject).NotTo(BeEmpty())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test creating token with valid subject
	token, err := provider.NewToken(user.Subject, 24*time.Hour)
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
	_, err = provider.NewToken("non-existent-subject", 24*time.Hour)
	g.Expect(err).NotTo(BeNil())

	// Test authenticating with the token
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify token claims
	claims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(claims[ClaimSub]).To(Equal(user.Subject))

	// Test that expired keys are rejected
	expiredSecret := "expired-secret-key"
	expiredKey := &model.Token{
		UserID: &user.ID,
	}
	err = db.Create(expiredKey).Error
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + expiredSecret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestTaskGrant tests creating and authenticating with task tokens.
func TestTaskGrant(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task in Running state (required for cache to load it)
	task := &Task{
		ID: 445,
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test creating task token
	token, err := provider.TaskGrant(task)
	g.Expect(err).To(BeNil())
	g.Expect(token.Secret).NotTo(BeEmpty())
	g.Expect(token.TaskID).NotTo(BeNil())
	g.Expect(*token.TaskID).To(Equal(uint(445)))

	// Test authenticating with the task token
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify token claims contain task subject (hex format)
	claims := jwToken.Claims.(jwt.MapClaims)
	subject := claims[ClaimSub].(string)
	expectedSubject := Task{ID: 445}.Subject()
	g.Expect(subject).To(Equal(expectedSubject))

	// Verify User() returns task login (decimal format)
	user := provider.User(jwToken)
	g.Expect(user).To(Equal(Task{ID: 445}.Login()))

	// Verify scopes are AddonScopes
	scopes := provider.Scopes(jwToken)
	g.Expect(scopes).NotTo(BeEmpty())
	scopeStrings := make([]string, len(scopes))
	for i, s := range scopes {
		scopeStrings[i] = s.String()
	}
	g.Expect(scopeStrings).To(ContainElement("addons:get"))
	g.Expect(scopeStrings).To(ContainElement("applications:get"))

	// Verify the token was created in the database
	var dbToken model.Token
	err = db.First(&dbToken, token.ID).Error
	g.Expect(err).To(BeNil())
	g.Expect(dbToken.TaskID).NotTo(BeNil())
	g.Expect(*dbToken.TaskID).To(Equal(uint(445)))
}

// TestTaskSubjectFormat tests that task subjects are formatted correctly.
func TestTaskSubjectFormat(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create tasks with different IDs
	task1 := &Task{ID: 1}
	task2 := &Task{ID: 999}
	task3 := &Task{ID: 12345}

	err = db.Create(task1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(task2).Error
	g.Expect(err).To(BeNil())
	err = db.Create(task3).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Test task.1 - Subject() returns hex format, User() returns login (decimal)
	token1, err := provider.TaskGrant(task1)
	g.Expect(err).To(BeNil())
	request := newTestRequest()
	request.With("Bearer " + token1.Secret)
	jwToken1, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	claims := jwToken1.Claims.(jwt.MapClaims)
	expectedSubject1 := Task{ID: 1}.Subject()
	g.Expect(claims[ClaimSub]).To(Equal(expectedSubject1))
	g.Expect(provider.User(jwToken1)).To(Equal(Task{ID: 1}.Login()))

	// Test task.999
	token2, err := provider.TaskGrant(task2)
	g.Expect(err).To(BeNil())
	request = newTestRequest()
	request.With("Bearer " + token2.Secret)
	jwToken2, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	claims = jwToken2.Claims.(jwt.MapClaims)
	expectedSubject2 := Task{ID: 999}.Subject()
	g.Expect(claims[ClaimSub]).To(Equal(expectedSubject2))
	g.Expect(provider.User(jwToken2)).To(Equal(Task{ID: 999}.Login()))

	// Test task.12345
	token3, err := provider.TaskGrant(task3)
	g.Expect(err).To(BeNil())
	request = newTestRequest()
	request.With("Bearer " + token3.Secret)
	jwToken3, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	claims = jwToken3.Claims.(jwt.MapClaims)
	expectedSubject3 := Task{ID: 12345}.Subject()
	g.Expect(claims[ClaimSub]).To(Equal(expectedSubject3))
	g.Expect(provider.User(jwToken3)).To(Equal(Task{ID: 12345}.Login()))
}

// TestTaskSubjectParsing tests that task subjects are parsed on-demand.
func TestTaskSubjectParsing(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create task
	task := &Task{ID: 100}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Grant token
	token, err := provider.TaskGrant(task)
	g.Expect(err).To(BeNil())

	// Verify task subject is parsed on-demand (not cached)
	expectedSubject := task.Subject()
	subject, err := provider.cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
	g.Expect(subject.Task).NotTo(BeNil())
	g.Expect(subject.Task.ID).To(Equal(uint(100)))
	g.Expect(subject.Login()).To(Equal(task.Login()))

	// Verify authentication works
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())
}

// TestTaskSubjectScopes tests that task subjects get AddonScopes.
func TestTaskSubjectScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	task := &Task{ID: 200}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	token, err := provider.TaskGrant(task)
	g.Expect(err).To(BeNil())

	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Verify scopes match AddonScopes
	scopes := provider.Scopes(jwToken)
	scopeStrings := make([]string, len(scopes))
	for i, s := range scopes {
		scopeStrings[i] = s.String()
	}

	// Check some key addon scopes
	g.Expect(scopeStrings).To(ContainElement("addons:get"))
	g.Expect(scopeStrings).To(ContainElement("applications:get"))
	g.Expect(scopeStrings).To(ContainElement("applications:post"))
	g.Expect(scopeStrings).To(ContainElement("applications:put"))
	g.Expect(scopeStrings).To(ContainElement("applications.facts:*"))
	g.Expect(scopeStrings).To(ContainElement("tasks:get"))
	g.Expect(scopeStrings).To(ContainElement("identities:get"))
}

// TestTaskTokenLifecycle tests the full lifecycle of task tokens.
func TestTaskTokenLifecycle(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create task
	task := &Task{ID: 300}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Task subject is always parseable (not cached)
	expectedSubject := task.Subject()
	subject, err := provider.cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
	g.Expect(subject.Task.ID).To(Equal(uint(300)))

	// Grant token
	token, err := provider.TaskGrant(task)
	g.Expect(err).To(BeNil())

	// Authentication works with token
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Revoke task - removes tokens (but subject still parseable)
	provider.TaskRevoke(300)

	// Subject still parseable
	subject, err = provider.cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())

	// Authentication fails (token removed)
	request = newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestMultipleTasksWithTokens tests multiple tasks can have tokens simultaneously.
func TestMultipleTasksWithTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create multiple tasks
	tasks := []*Task{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
	for _, task := range tasks {
		err = db.Create(task).Error
		g.Expect(err).To(BeNil())
	}

	// Grant tokens for all tasks
	tokens := make([]Token, len(tasks))
	for i, task := range tasks {
		tokens[i], err = provider.TaskGrant(task)
		g.Expect(err).To(BeNil())
	}

	// All task subjects are parseable and authenticate with tokens
	for i, task := range tasks {
		expectedSubject := task.Subject()

		// Subject is always parseable (not cached)
		subject, err := provider.cache.FindSubject(expectedSubject)
		g.Expect(err).To(BeNil())
		g.Expect(subject.IsTask()).To(BeTrue())
		g.Expect(subject.Login()).To(Equal(task.Login()))

		// Check authentication
		request := newTestRequest()
		request.With("Bearer " + tokens[i].Secret)
		jwToken, err := provider.Authenticate(request)
		g.Expect(err).To(BeNil())
		claims := jwToken.Claims.(jwt.MapClaims)
		g.Expect(claims[ClaimSub]).To(Equal(expectedSubject))
		g.Expect(provider.User(jwToken)).To(Equal(task.Login()))
	}

	// Revoke one task - others still work
	provider.TaskRevoke(2)

	// Task 2 subject still parseable
	task2Subject := Task{ID: 2}.Subject()
	subject, err := provider.cache.FindSubject(task2Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())

	// Task 2 token authentication fails (token removed)
	request := newTestRequest()
	request.With("Bearer " + tokens[1].Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Task 1 and 3 still work
	request = newTestRequest()
	request.With("Bearer " + tokens[0].Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + tokens[2].Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())
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
	claims[ClaimIss] = testIssuer()
	claims[ClaimIat] = time.Now().Unix()

	tokenString, err := token.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	// Test authenticating with valid JWT
	request := newTestRequest()
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
	expiredClaims[ClaimIss] = testIssuer()
	expiredClaims[ClaimIat] = time.Now().Add(-2 * time.Hour).Unix()

	expiredTokenString, err := expiredToken.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + expiredTokenString)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("expired"))

	// Test with missing sub claim
	noSubToken := jwt.New(jwt.SigningMethodHS512)
	noSubClaims := noSubToken.Claims.(jwt.MapClaims)
	noSubClaims[ClaimScope] = "openid"
	noSubClaims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	noSubClaims[ClaimIss] = testIssuer()
	noSubClaims[ClaimIat] = time.Now().Unix()

	noSubTokenString, err := noSubToken.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + noSubTokenString)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("User not specified"))

	// Test with missing scope claim
	noScopeToken := jwt.New(jwt.SigningMethodHS512)
	noScopeClaims := noScopeToken.Claims.(jwt.MapClaims)
	noScopeClaims[ClaimSub] = "user-123"
	noScopeClaims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	noScopeClaims[ClaimIss] = testIssuer()
	noScopeClaims[ClaimIat] = time.Now().Unix()

	noScopeTokenString, err := noScopeToken.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + noScopeTokenString)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("Scope not specified"))
}

// TestLegacyHMACTokenAuthentication tests authenticating legacy HMAC tokens
// that were created before iss claim validation was added (simulates old NewToken() behavior).
func TestLegacyHMACTokenAuthentication(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Simulate old NewToken() behavior:
	// - HMAC signing
	// - Sets sub, scope, exp
	// - NO iss claim (this is what we're testing)
	// - Custom claims (like task ID)
	signingKey := []byte(Settings.Auth.Token.Key)

	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims[ClaimSub] = "addon-user-123"
	claims[ClaimScope] = "tasks:post applications:get"
	claims[ClaimExp] = float64(time.Now().Add(24 * time.Hour).Unix())
	// NO iss claim - simulates legacy token
	claims["task"] = 42 // Custom claim from old NewToken() usage

	tokenString, err := token.SignedString(signingKey)
	g.Expect(err).To(BeNil())

	// Authenticate with legacy token
	request := newTestRequest()
	request.With("Bearer " + tokenString)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify claims
	jwtClaims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(jwtClaims[ClaimSub]).To(Equal("addon-user-123"))
	g.Expect(jwtClaims[ClaimScope]).To(Equal("tasks:post applications:get"))
	g.Expect(jwtClaims["task"]).To(Equal(float64(42)))

	// Verify iss was injected from request context
	g.Expect(jwtClaims[ClaimIss]).To(Equal(testIssuer()))
}

// TestUserExtraction tests extracting user from token.
func TestUserExtraction(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create a user in the database
	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "testuser456",
		Password: secret.HashPassword("password"),
		Email:    "testuser456@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get auto-generated subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create a token with claims
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims[ClaimSub] = user.Subject
	claims[ClaimScope] = "openid profile"
	claims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	claims[ClaimIss] = testIssuer()
	claims[ClaimIat] = time.Now().Unix()

	login := provider.User(token)
	g.Expect(login).To(Equal("testuser456"))

	// Test with missing sub claim
	tokenNoSub := jwt.New(jwt.SigningMethodHS256)
	noSubClaims := tokenNoSub.Claims.(jwt.MapClaims)
	noSubClaims[ClaimScope] = "openid"
	noSubClaims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	noSubClaims[ClaimIss] = testIssuer()
	noSubClaims[ClaimIat] = time.Now().Unix()

	login = provider.User(tokenNoSub)
	g.Expect(login).To(BeEmpty())
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
	claims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	claims[ClaimIss] = testIssuer()
	claims[ClaimIat] = time.Now().Unix()

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
	noScopeClaims := tokenNoScope.Claims.(jwt.MapClaims)
	noScopeClaims[ClaimSub] = "test-user"
	noScopeClaims[ClaimExp] = float64(time.Now().Add(1 * time.Hour).Unix())
	noScopeClaims[ClaimIss] = testIssuer()
	noScopeClaims[ClaimIat] = time.Now().Unix()

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
	request := newTestRequest()
	request.With("invalid-token-without-bearer")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test empty token
	request = newTestRequest()
	request.With("")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test malformed bearer token (missing token value)
	request = newTestRequest()
	request.With("Bearer")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test invalid JWT
	request = newTestRequest()
	request.With("Bearer invalid.jwt.token")
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestBaseScopeMatching tests scope matching with wildcards and exact matches.
func TestBaseScopeMatching(t *testing.T) {
	g := NewGomegaWithT(t)

	// Wildcard scope matches everything
	scope := Scope{Resource: "*", Method: "*"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("tags", "POST")).To(BeTrue())

	// Resource wildcard matches any method for that resource
	scope = Scope{Resource: "applications", Method: "*"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeTrue())
	g.Expect(scope.Match("tags", "GET")).To(BeFalse())

	// Method wildcard matches that method for any resource
	scope = Scope{Resource: "*", Method: "GET"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("tags", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeFalse())

	// Exact match
	scope = Scope{Resource: "applications", Method: "GET"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeFalse())
	g.Expect(scope.Match("tags", "GET")).To(BeFalse())

	// Case insensitive
	scope = Scope{Resource: "Applications", Method: "get"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
}

// TestBaseScopeParsing tests parsing scope strings.
func TestBaseScopeParsing(t *testing.T) {
	g := NewGomegaWithT(t)

	scope := Scope{}
	scope.With("applications:read")
	g.Expect(scope.Resource).To(Equal("applications"))
	g.Expect(scope.Method).To(Equal("read"))

	scope = Scope{}
	scope.With("*:*")
	g.Expect(scope.Resource).To(Equal("*"))
	g.Expect(scope.Method).To(Equal("*"))

	// Test String() roundtrip
	scope = Scope{Resource: "tags", Method: "write"}
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
	taskRef := &Task{ID: task.ID}
	key, err := provider.TaskGrant(taskRef)
	g.Expect(err).To(BeNil())

	// Authenticate with key - should work
	request := newTestRequest()
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Update task to Succeeded - key should now be rejected
	provider.cache.Reset()
	db.Model(task).Update("State", "Succeeded")
	request = newTestRequest()
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))

	// Test with Failed state
	provider.cache.Reset()
	db.Model(task).Update("State", "Failed")
	request = newTestRequest()
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	// Test with Canceled state
	provider.cache.Reset()
	db.Model(task).Update("State", "Canceled")
	request = newTestRequest()
	request.With("Bearer " + key.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
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
	request := newTestRequest()
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
	login := provider.User(token)
	g.Expect(login).To(Equal("admin.noauth"))

	// Create test user.
	user := &model.User{
		Subject:  "noauth-test-subject",
		Login:    "testuser",
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

	key, err := provider.NewToken(user.Subject, time.Hour)
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

	taskRef := &Task{ID: task.ID}
	key, err = provider.TaskGrant(taskRef)
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
		TokenId: "test-token-id",
		Reason:  "expired",
	}
	g.Expect(err.Error()).To(ContainSubstring("test-token-id"))
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
		Login:    "cachedeleteuser",
		Password: secret.HashPassword("password"),
		Email:    "cachedelete@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Create provider with cache
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token
	token, err := provider.NewToken(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.ID).NotTo(Equal(uint(0)))

	// Verify token is in cache by authenticating
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Delete token from DB first to prevent refresh from reloading it
	err = db.Delete(&model.Token{}, token.ID).Error
	g.Expect(err).To(BeNil())

	// Delete token using cache method
	provider.cache.TokenDeleted(token.ID)

	// Verify token is removed from both indexes by checking authentication fails
	request = newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))
}

// TestTaskRevoke tests revoking task tokens.
func TestTaskRevoke(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task
	task := &model.Task{
		Name:  "revoke-test-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create task token
	taskRef := &Task{ID: task.ID}
	token, err := provider.TaskGrant(taskRef)
	g.Expect(err).To(BeNil())

	// Verify token works
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Verify token exists in database
	var count int64
	db.Model(&model.Token{}).Where("TaskID = ?", task.ID).Count(&count)
	g.Expect(count).To(Equal(int64(1)))

	// Revoke task token
	provider.TaskRevoke(task.ID)

	// Verify token is deleted from database
	db.Model(&model.Token{}).Where("TaskID = ?", task.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Verify token no longer authenticates (removed from cache)
	request = newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not-authenticated"))
}

// TestTaskRevokeMultipleTokens tests revoking when task has multiple tokens (edge case).
func TestTaskRevokeMultipleTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task
	task := &model.Task{
		Name:  "multi-token-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create first token
	taskRef := &Task{ID: task.ID}
	token1, err := provider.TaskGrant(taskRef)
	g.Expect(err).To(BeNil())

	// Manually create second token for same task (shouldn't normally happen)
	token2Secret := "second-task-token"
	taskCache := &Task{ID: task.ID}
	token2 := &model.Token{
		Kind:       KindAPIKey,
		Subject:    taskCache.Subject(),
		AuthId:     "second-auth-id",
		Digest:     secret.Hash(token2Secret),
		Expiration: time.Now().Add(24 * time.Hour),
		TaskID:     &task.ID,
	}
	err = db.Create(token2).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to pick up both tokens
	err = provider.cache.Refresh()
	g.Expect(err).To(BeNil())

	// Verify both tokens work
	request := newTestRequest()
	request.With("Bearer " + token1.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + token2Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Revoke all tokens for this task
	provider.TaskRevoke(task.ID)

	// Verify both tokens are deleted from database
	var count int64
	db.Model(&model.Token{}).Where("TaskID = ?", task.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Verify neither token authenticates
	request = newTestRequest()
	request.With("Bearer " + token1.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())

	request = newTestRequest()
	request.With("Bearer " + token2Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil())
}

// TestTaskRevokeNoTokens tests revoking task with no tokens (should not error).
func TestTaskRevokeNoTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test task
	task := &model.Task{
		Name:  "no-token-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Revoke should not error even though no tokens exist
	provider.TaskRevoke(task.ID)

	// Verify no tokens exist for this task
	var count int64
	db.Model(&model.Token{}).Where("TaskID = ?", task.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Note: Task may be reloaded into cache by ensureFresh() since it's still
	// in Running state in the database. This is expected behavior.
}

// TestCascadeDeleteUser tests that deleting a user cascades to delete tokens.
func TestCascadeDeleteUser(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create user
	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "cascadeuser",
		Password: secret.HashPassword("password"),
		Email:    "cascade@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token for user
	token, err := provider.NewToken(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Verify token exists
	var count int64
	db.Model(&model.Token{}).Where("UserID = ?", user.ID).Count(&count)
	g.Expect(count).To(Equal(int64(1)))

	// Delete user (should cascade delete token)
	err = db.Delete(user).Error
	g.Expect(err).To(BeNil())

	// Verify token was cascade deleted
	db.Model(&model.Token{}).Where("UserID = ?", user.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Verify token no longer in database by ID
	var deletedToken model.Token
	err = db.First(&deletedToken, token.ID).Error
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.Is(err, gorm.ErrRecordNotFound)).To(BeTrue())
}

// TestCascadeDeleteTask tests that deleting a task cascades to delete tokens.
func TestCascadeDeleteTask(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create task
	task := &model.Task{
		Name:  "cascade-task",
		State: "Running",
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token for task
	taskRef := &Task{ID: task.ID}
	token, err := provider.TaskGrant(taskRef)
	g.Expect(err).To(BeNil())

	// Verify token exists
	var count int64
	db.Model(&model.Token{}).Where("TaskID = ?", task.ID).Count(&count)
	g.Expect(count).To(Equal(int64(1)))

	// Delete task (should cascade delete token)
	err = db.Delete(task).Error
	g.Expect(err).To(BeNil())

	// Verify token was cascade deleted
	db.Model(&model.Token{}).Where("TaskID = ?", task.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Verify token no longer in database by ID
	var deletedToken model.Token
	err = db.First(&deletedToken, token.ID).Error
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.Is(err, gorm.ErrRecordNotFound)).To(BeTrue())
}

// TestClientPAT tests creating and authenticating with client PAT.
func TestClientPAT(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create IdP client
	client := &model.IdpClient{
		Subject:         uuid.New().String(),
		ClientId:        "test-client",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid", "profile"},
	}
	err = db.Create(client).Error
	g.Expect(err).To(BeNil())
	g.Expect(client.Subject).NotTo(BeEmpty())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create PAT for client
	token, err := provider.NewToken(client.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.Secret).NotTo(BeEmpty())
	g.Expect(token.IdpClientID).NotTo(BeNil())
	g.Expect(*token.IdpClientID).To(Equal(client.ID))

	// Authenticate with the token
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify token claims contain client subject and scopes
	claims := jwToken.Claims.(jwt.MapClaims)
	g.Expect(claims[ClaimSub]).To(Equal(client.Subject))
	scopeStr := claims[ClaimScope].(string)
	g.Expect(scopeStr).To(ContainSubstring("openid"))
	g.Expect(scopeStr).To(ContainSubstring("profile"))
}

// TestCascadeDeleteIdpClient tests that deleting an IdP client cascades to delete tokens.
func TestCascadeDeleteIdpClient(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create IdP client
	client := &model.IdpClient{
		Subject:         uuid.New().String(),
		ClientId:        "cascade-client",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	err = db.Create(client).Error
	g.Expect(err).To(BeNil())
	g.Expect(client.Subject).NotTo(BeEmpty())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token for client (simulating client credentials grant)
	token, err := provider.NewToken(client.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.IdpClientID).NotTo(BeNil())
	g.Expect(*token.IdpClientID).To(Equal(client.ID))

	// Verify token exists
	var count int64
	db.Model(&model.Token{}).Where("IdpClientID = ?", client.ID).Count(&count)
	g.Expect(count).To(Equal(int64(1)))

	// Delete client (should cascade delete token)
	err = db.Delete(client).Error
	g.Expect(err).To(BeNil())

	// Verify token was cascade deleted
	db.Model(&model.Token{}).Where("IdpClientID = ?", client.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Verify token no longer in database by ID
	var deletedToken model.Token
	err = db.First(&deletedToken, token.ID).Error
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.Is(err, gorm.ErrRecordNotFound)).To(BeTrue())
}

// TestCascadeDeleteIdpIdentity tests that deleting an IdP identity cascades to delete tokens.
func TestCascadeDeleteIdpIdentity(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create IdP identity
	identity := &Identity{
		Issuer:  "https://cascade.idp.com",
		Subject: "cascade-idp-subject",
		Login:   "cascadeidentity",
		Email:   "cascadeidentity@example.com",
	}
	err = secret.Encrypt(identity)
	g.Expect(err).To(BeNil())
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token for identity
	token, err := provider.NewToken(identity.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Verify token exists
	var count int64
	db.Model(&model.Token{}).Where("IdpIdentityID = ?", identity.ID).Count(&count)
	g.Expect(count).To(Equal(int64(1)))

	// Delete identity (should cascade delete token)
	err = db.Delete(identity).Error
	g.Expect(err).To(BeNil())

	// Verify token was cascade deleted
	db.Model(&model.Token{}).Where("IdpIdentityID = ?", identity.ID).Count(&count)
	g.Expect(count).To(Equal(int64(0)))

	// Verify token no longer in database by ID
	var deletedToken model.Token
	err = db.First(&deletedToken, token.ID).Error
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.Is(err, gorm.ErrRecordNotFound)).To(BeTrue())
}

// TestBuiltinRevoke tests the Builtin Revoke method removes key from cache and DB.
func TestBuiltinRevoke(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user with scopes
	user := &model.User{
		Subject:  "builtin-delete-user",
		Login:    "builtindeleteuser",
		Password: secret.HashPassword("password"),
		Email:    "builtindelete@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	role := &model.Role{
		Name:   "Admin",
		Scopes: []string{"*:*"},
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	err = db.Model(user).Association("Roles").Append(role)
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token
	token, err := provider.NewToken(user.Subject, 1*time.Hour)
	g.Expect(err).To(BeNil())

	// Populate cache by authenticating
	request := newTestRequest()
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
		Login:    "cachenotifuser",
		Password: secret.HashPassword("password"),
		Email:    "cachenotif@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// User should be in cache (loaded during provider initialization)
	token, err := provider.NewToken(user.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Verify token works (NewToken calls TokenSaved notification)
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Delete token from DB and notify cache
	err = db.Delete(&model.Token{}, token.ID).Error
	g.Expect(err).To(BeNil())
	provider.cache.TokenDeleted(token.ID)

	// Verify token is immediately gone from cache (notification propagated)
	request = newTestRequest()
	request.With("Bearer " + token.Secret)
	_, err = provider.Authenticate(request)
	g.Expect(err).NotTo(BeNil()) // Should fail immediately, not wait for refresh
}

// TestIdpIdentityTokenBinding tests token binding to IdP identities.
func TestIdpIdentityTokenBinding(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create IdP identity
	identity := &Identity{
		Issuer:  "https://idp.example.com",
		Subject: "idp-user-123",
		Login:   "idpuser",
		Email:   "idpuser@example.com",
		Scopes:  []string{"openid profile email"},
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token bound to IdP identity (scopes come from grant)
	token, err := provider.NewToken(identity.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())
	g.Expect(token.IdpIdentityID).NotTo(BeNil())
	g.Expect(*token.IdpIdentityID).To(Equal(identity.ID))

	// Authenticate with IdP-bound token
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify claims - scopes injected from grant
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
		Login:    "usernoroles",
		Password: secret.HashPassword("password"),
		Email:    "noroles@example.com",
	}
	err = db.Create(userNoRoles).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	token, err := provider.NewToken(userNoRoles.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	claims := jwToken.Claims.(jwt.MapClaims)
	scopes := claims[ClaimScope].(string)
	g.Expect(scopes).To(BeEmpty())

	// Test 2: Role with no scopes
	roleNoPerm := &model.Role{
		Name: "EmptyRole",
	}
	err = db.Create(roleNoPerm).Error
	g.Expect(err).To(BeNil())

	userEmptyRole := &model.User{
		Subject:  "user-empty-role",
		Login:    "useremptyrole",
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

	token, err = provider.NewToken(userEmptyRole.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err = provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	claims = jwToken.Claims.(jwt.MapClaims)
	scopes = claims[ClaimScope].(string)
	g.Expect(scopes).To(BeEmpty())

	// Test 3: Multiple roles with overlapping scopes (deduplication)
	role1 := &model.Role{
		Name:   "Reader",
		Scopes: []string{"apps:GET"},
	}
	role2 := &model.Role{
		Name:   "Writer",
		Scopes: []string{"apps:GET", "apps:POST"}, // apps:GET overlaps with role1
	}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(role2).Error
	g.Expect(err).To(BeNil())

	userMultiRole := &model.User{
		Subject:  "user-multi-role",
		Login:    "usermultirole",
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

	token, err = provider.NewToken(userMultiRole.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	request = newTestRequest()
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

	taskRef := &Task{ID: pendingTask.ID}
	token, err := provider.TaskGrant(taskRef)
	g.Expect(err).To(BeNil())

	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())
	g.Expect(jwToken).NotTo(BeNil())

	// Verify task subject format
	claims := jwToken.Claims.(jwt.MapClaims)
	subject := claims[ClaimSub].(string)
	g.Expect(subject).To(Equal(taskRef.Subject()))
}

// TestCacheFindSubject tests finding subjects (users and identities) by subject string.
func TestCacheFindSubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user with roles and scopes
	role := &model.Role{
		Name:   "AppReader",
		Scopes: []string{"applications:GET"},
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "testuser",
		Password: secret.HashPassword("password"),
		Email:    "user@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get auto-generated subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())

	err = db.Model(user).Association("Roles").Append(role)
	g.Expect(err).To(BeNil())

	// Create test IdP identity
	identity := &Identity{
		Issuer:  "https://idp.example.com",
		Subject: "idp-subject-456",
		Email:   "idp@example.com",
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
	g.Expect(subject.Key).To(Equal(user.Subject))
	g.Expect(subject.User.Login).To(Equal("testuser"))
	g.Expect(subject.Email).To(Equal("user@example.com"))
	g.Expect(subject.Scopes).To(ContainElement("applications:GET"))

	// Test finding identity by subject
	subject, err = provider.cache.FindSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsUser()).To(BeFalse())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.Key).To(Equal("idp-subject-456"))
	g.Expect(subject.Identity.Login).To(Equal(""))
	g.Expect(subject.Email).To(Equal("idp@example.com"))
	// IdpIdentity no longer has scopes - they come from Grant or Token
	g.Expect(subject.Scopes).To(BeEmpty())
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
		Login:    "newuser",
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

// TestCacheUserSavedBySubject tests that UserSaved updates bySubject map.
func TestStorageFindSubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test data
	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "storageuser",
		Password: secret.HashPassword("password"),
		Email:    "storage@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get auto-generated subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())

	identity := &Identity{
		Issuer:  "https://storage.idp.com",
		Subject: "storage-identity-subject",
		Login:   "storageidentity",
		Email:   "storageidentity@example.com",
	}
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
	g.Expect(subject.Key).To(Equal(user.Subject))
	g.Expect(subject.User.Login).To(Equal("storageuser"))

	// Find identity subject
	subject, err = storage.findSubject(identity.Subject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsIdentity()).To(BeTrue())
	g.Expect(subject.Key).To(Equal("storage-identity-subject"))
	g.Expect(subject.Identity.Login).To(Equal("storageidentity"))

	// Find non-existent subject
	_, err = storage.findSubject("non-existent")
	g.Expect(err).NotTo(BeNil())
}

// TestCacheFindUserByLogin tests finding user by userid field.
func TestCacheFindUserByLogin(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test user
	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "testuserid",
		Password: secret.HashPassword("password"),
		Email:    "userid@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get auto-generated subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Find by userid
	found, err := provider.cache.FindUserByLogin("testuserid")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(Equal(user.Subject))
	g.Expect(found.Email).To(Equal("userid@example.com"))

	// Find non-existent userid
	_, err = provider.cache.FindUserByLogin("nonexistent")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("user"))
}

// TestCacheFindUserByLoginNotification tests notification-based cache updates.
func TestCacheFindUserByLoginNotification(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Force initial cache load
	_, _ = provider.cache.FindUserByLogin("force-initial-load")

	// Create user after cache is loaded (NOT notified)
	user := &model.User{
		Subject:  uuid.New().String(),
		Login:    "newuserid",
		Password: secret.HashPassword("password"),
		Email:    "newuserid@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Reload user to get auto-generated subject
	err = db.First(user, user.ID).Error
	g.Expect(err).To(BeNil())

	// FindUserByLogin should NOT find it (cache is fresh, no notification)
	found, err := provider.cache.FindUserByLogin("newuserid")
	g.Expect(err).NotTo(BeNil()) // NotFound
	g.Expect(found).To(BeNil())

	// Notify cache of user creation
	provider.cache.UserSaved((*User)(user))

	// Now it should be found immediately
	found, err = provider.cache.FindUserByLogin("newuserid")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(Equal(user.Subject))
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

// TestSeedClientsFromCRD tests seeding clients from CRDs in disconnected mode.
func TestSeedClientsFromCRD(t *testing.T) {
	g := NewGomegaWithT(t)

	// Setup disconnected mode
	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
		federated = &as.Federated{} // Reset federated for next test
	}()
	Settings.Disconnected = true

	// Setup DB
	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Load federated settings (gets fake client with seed data)
	err = federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	// Seed clients
	domain := NewTenant(db)
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify database state
	var clients []IdpClient
	err = db.Find(&clients).Error
	g.Expect(err).To(BeNil())
	g.Expect(clients).To(HaveLen(4))

	// Find and verify web-ui client
	var webUI IdpClient
	err = db.First(&webUI, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(webUI.ClientId).To(Equal("web-ui"))
	g.Expect(webUI.ApplicationType).To(Equal("web"))
	g.Expect(webUI.Secret).To(BeEmpty())
	g.Expect(webUI.Grants).To(ContainElement("authorization_code"))
	g.Expect(webUI.Scopes).To(ContainElement("openid"))

	// Find and verify kantra client
	var kantra IdpClient
	err = db.First(&kantra, "ClientId = ?", "kantra").Error
	g.Expect(err).To(BeNil())
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kantra.ApplicationType).To(Equal("native"))
	g.Expect(kantra.Secret).To(BeEmpty())

	// Find and verify kai-ide client
	var kaiIDE IdpClient
	err = db.First(&kaiIDE, "ClientId = ?", "kai-ide").Error
	g.Expect(err).To(BeNil())
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))
	g.Expect(kaiIDE.ApplicationType).To(Equal("native"))
	g.Expect(kaiIDE.Secret).To(BeEmpty())

	// Find and verify web-ui-with-secret client (confidential)
	var confidential IdpClient
	err = db.First(&confidential, "ClientId = ?", "web-ui-with-secret").Error
	g.Expect(err).To(BeNil())
	g.Expect(confidential.ID).To(Equal(uint(4)))
	g.Expect(confidential.ApplicationType).To(Equal("web"))
	g.Expect(confidential.Secret).NotTo(BeEmpty())
}

// TestSeedClientsUpdate tests updating existing clients from CRDs.
func TestSeedClientsUpdate(t *testing.T) {
	g := NewGomegaWithT(t)

	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
		federated = &as.Federated{} // Reset federated for next test
	}()
	Settings.Disconnected = true

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Pre-create client in DB with different grants
	existing := &IdpClient{
		ClientId:        "web-ui",
		ApplicationType: "web",
		Grants:          []string{"old-grant"},
		Scopes:          []string{"openid"},
	}
	existing.ID = 1
	err = db.Create(existing).Error
	g.Expect(err).To(BeNil())

	// Load CRDs
	err = federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	// Seed clients
	domain := NewTenant(db)
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify client was updated (not recreated)
	var updated IdpClient
	err = db.First(&updated, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	g.Expect(updated.ID).To(Equal(uint(1))) // ID preserved
	g.Expect(updated.Grants).To(ContainElement("authorization_code"))
	g.Expect(updated.Grants).To(ContainElement("refresh_token"))
	g.Expect(updated.Grants).NotTo(ContainElement("old-grant"))
	g.Expect(updated.Secret).To(BeEmpty())

	// Verify confidential client has secret resolved
	var confidential IdpClient
	err = db.First(&confidential, "ClientId = ?", "web-ui-with-secret").Error
	g.Expect(err).To(BeNil())
	g.Expect(confidential.Secret).NotTo(BeEmpty())
}

// TestSeedClientsDeleteOrphaned tests deleting orphaned seeded clients.
func TestSeedClientsDeleteOrphaned(t *testing.T) {
	g := NewGomegaWithT(t)

	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
		federated = &as.Federated{} // Reset federated for next test
	}()
	Settings.Disconnected = true

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create orphaned seeded client (ID < 1000, not in CRDs)
	orphaned := &IdpClient{
		ClientId:        "orphaned-client",
		ApplicationType: "web",
		Grants:          []string{"authorization_code"},
		Scopes:          []string{"openid"},
	}
	orphaned.ID = 500
	err = db.Create(orphaned).Error
	g.Expect(err).To(BeNil())

	// Create non-seeded client (ID >= 1000, should be preserved)
	nonSeeded := &IdpClient{
		ClientId:        "custom-client",
		ApplicationType: "native",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"api"},
	}
	nonSeeded.ID = 1001
	err = db.Create(nonSeeded).Error
	g.Expect(err).To(BeNil())

	// Load CRDs (web-ui, kantra, kai-ide)
	err = federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	// Seed clients
	domain := NewTenant(db)
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

	// Verify total count (4 from CRDs + 1 custom)
	var count int64
	db.Model(&IdpClient{}).Count(&count)
	g.Expect(count).To(Equal(int64(5)))
}

// TestSeedClientsIDPreservation tests that IDs from CRDs are preserved across multiple seeds.
func TestSeedClientsIDPreservation(t *testing.T) {
	g := NewGomegaWithT(t)

	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
		federated = &as.Federated{} // Reset federated for next test
	}()
	Settings.Disconnected = true

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Load CRDs
	err = federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	domain := NewTenant(db)

	// First seed
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify IDs after first seed
	var webUI, kantra, kaiIDE IdpClient
	err = db.First(&webUI, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	err = db.First(&kantra, "ClientId = ?", "kantra").Error
	g.Expect(err).To(BeNil())
	err = db.First(&kaiIDE, "ClientId = ?", "kai-ide").Error
	g.Expect(err).To(BeNil())

	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))

	// Verify count after first seed
	var count int64
	db.Model(&IdpClient{}).Count(&count)
	g.Expect(count).To(Equal(int64(4)))

	// Second seed (should preserve IDs)
	err = domain.seedClients(db)
	g.Expect(err).To(BeNil())

	// Verify IDs after second seed
	err = db.First(&webUI, "ClientId = ?", "web-ui").Error
	g.Expect(err).To(BeNil())
	err = db.First(&kantra, "ClientId = ?", "kantra").Error
	g.Expect(err).To(BeNil())
	err = db.First(&kaiIDE, "ClientId = ?", "kai-ide").Error
	g.Expect(err).To(BeNil())

	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))

	// Verify count after second seed (should still be 4)
	db.Model(&IdpClient{}).Count(&count)
	g.Expect(count).To(Equal(int64(4)))
}

// TestClientWithWildcard tests Client.With() with wildcard redirect URIs.
func TestClientWithWildcard(t *testing.T) {
	g := NewGomegaWithT(t)

	// Test exact wildcard match
	m := &IdpClient{
		ClientId:        "test-client",
		Subject:         "test-subject",
		Secret:          "test-secret",
		ApplicationType: "web",
		Grants:          []string{"authorization_code"},
		RedirectURIs:    []string{"http://*.example.com/callback"},
		Scopes:          []string{"openid"},
	}

	req := httptest.NewRequest("GET", "http://localhost:8080/oidc", nil)
	client := &Client{}
	client.With(m, req)

	g.Expect(client.id).To(Equal("test-client"))
	g.Expect(client.subject).To(Equal("test-subject"))
	g.Expect(client.secret).To(Equal("test-secret"))
	g.Expect(client.applicationType).To(Equal(op.ApplicationTypeWeb))
	g.Expect(client.grantTypes).To(Equal([]string{"authorization_code"}))
	g.Expect(client.redirectURIs).To(Equal([]string{"http://*.example.com/callback"}))
	g.Expect(client.scopes).To(Equal([]string{"openid"}))
}

// TestClientWithMultipleRedirectURIs tests Client.With() with multiple redirect URIs.
func TestClientWithMultipleRedirectURIs(t *testing.T) {
	g := NewGomegaWithT(t)

	m := &IdpClient{
		ClientId:        "multi-redirect",
		ApplicationType: "web",
		Grants:          []string{"authorization_code", "refresh_token"},
		RedirectURIs: []string{
			"http://localhost:8080/callback",
			"https://*.prod.example.com/callback",
			"https://app.example.com/auth",
		},
		Scopes: []string{"openid", "profile", "email"},
	}

	req := httptest.NewRequest("GET", "http://localhost:8080/oidc", nil)
	client := &Client{}
	client.With(m, req)

	g.Expect(client.redirectURIs).To(HaveLen(3))
	g.Expect(client.redirectURIs[0]).To(Equal("http://localhost:8080/callback"))
	g.Expect(client.redirectURIs[1]).To(Equal("https://*.prod.example.com/callback"))
	g.Expect(client.redirectURIs[2]).To(Equal("https://app.example.com/auth"))
}

// TestClientInjectWildcardMatch tests Client.Inject() wildcard matching.
func TestClientInjectWildcardMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	// Wildcard pattern is matched against requested redirect URI
	issuer := "http://hub.example.com/oidc"
	requestedRedirect := "http://hub.example.com/callback"
	req := httptest.NewRequest(
		"GET",
		issuer+"/authorize?redirect_uri="+requestedRedirect,
		nil,
	)

	client := &Client{
		id:              "wildcard-client",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs:    []string{"http://*/callback"},
		request:         req,
	}

	client.Inject()

	// Wildcard pattern matched requested redirect - replaced with it
	g.Expect(client.redirectURIs).To(HaveLen(1))
	g.Expect(client.redirectURIs[0]).To(Equal(requestedRedirect))
}

// TestClientInjectWildcardNoMatch tests Client.Inject() when wildcard doesn't match requested URI.
func TestClientInjectWildcardNoMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	issuer := "http://hub.example.com/oidc"
	requestedRedirect := "http://different.example.com/callback"
	req := httptest.NewRequest(
		"GET",
		issuer+"/authorize?redirect_uri="+requestedRedirect,
		nil,
	)

	client := &Client{
		id:              "wildcard-mismatch",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs:    []string{"http://hub.example.com/*"},
		request:         req,
	}

	client.Inject()

	// Wildcard pattern doesn't match requested URI - pattern preserved
	g.Expect(client.redirectURIs).To(HaveLen(1))
	g.Expect(client.redirectURIs[0]).To(Equal("http://hub.example.com/*"))
}

// TestClientInjectMultipleWildcards tests Client.Inject() with multiple wildcard patterns.
func TestClientInjectMultipleWildcards(t *testing.T) {
	g := NewGomegaWithT(t)

	issuer := "https://prod.konveyor.io/hub/oidc"
	requestedRedirect := "https://prod.konveyor.io/hub/callback"
	req := httptest.NewRequest(
		"GET",
		issuer+"/authorize?redirect_uri="+requestedRedirect,
		nil,
	)

	client := &Client{
		id:              "multi-wildcard",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs: []string{
			"http://localhost:*/callback",        // Doesn't match (scheme mismatch)
			"https://*.konveyor.io/*/callback",   // Matches requested redirect
			"https://fixed.example.com/callback", // Fixed - no wildcard
		},
		request: req,
	}

	client.Inject()

	// First wildcard doesn't match requested redirect (http vs https)
	g.Expect(client.redirectURIs[0]).To(Equal("http://localhost:*/callback"))
	// Second wildcard matches requested redirect - replaced with it
	g.Expect(client.redirectURIs[1]).To(Equal(requestedRedirect))
	// Third is fixed - unchanged
	g.Expect(client.redirectURIs[2]).To(Equal("https://fixed.example.com/callback"))
}

// TestClientInjectTemplateVariables tests Client.Inject() template variable substitution.
func TestClientInjectTemplateVariables(t *testing.T) {
	g := NewGomegaWithT(t)

	req := httptest.NewRequest("GET", "http://hub.example.com:8080/oidc/authorize", nil)
	client := &Client{
		id:              "template-client",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs: []string{
			"${issuer}/callback",
			"http://${issuer.host}/auth",
			"http://localhost:${issuer.port}/callback",
			"http://example.com${issuer.path}/done",
		},
		request: req,
	}

	client.Inject()

	// ${issuer} should be replaced with full issuer
	g.Expect(client.redirectURIs[0]).To(Equal("http://hub.example.com:8080/oidc/callback"))
	// ${issuer.host} should be replaced with hostname (no port)
	g.Expect(client.redirectURIs[1]).To(Equal("http://hub.example.com/auth"))
	// ${issuer.port} should be replaced with port
	g.Expect(client.redirectURIs[2]).To(Equal("http://localhost:8080/callback"))
	// ${issuer.path} should be replaced with path
	g.Expect(client.redirectURIs[3]).To(Equal("http://example.com/oidc/done"))
}

// TestClientInjectCombinedWildcardAndTemplate tests wildcard and template together.
func TestClientInjectCombinedWildcardAndTemplate(t *testing.T) {
	g := NewGomegaWithT(t)

	issuer := "https://app.konveyor.io:443/hub/oidc"
	requestedRedirect := "https://app.konveyor.io:443/hub/callback"
	req := httptest.NewRequest(
		"GET",
		issuer+"/authorize?redirect_uri="+requestedRedirect,
		nil,
	)

	client := &Client{
		id:              "combined-client",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs: []string{
			"https://*.konveyor.io:*/hub/callback",    // Wildcard - matches requested
			"http://localhost:${issuer.port}/auth",    // Template - substitution only
			"https://fixed.example.com/callback",      // Fixed - unchanged
			"${issuer.proto}://${issuer.host}{,*}/**", // Wildcard - match host.
			"**", // Naked wildcard
		},
		request: req,
	}

	client.Inject()

	// Wildcard matches requested redirect - replaced with it
	g.Expect(client.redirectURIs[0]).To(Equal(requestedRedirect))
	// Template variables substituted
	g.Expect(client.redirectURIs[1]).To(Equal("http://localhost:443/auth"))
	// Fixed URI unchanged
	g.Expect(client.redirectURIs[2]).To(Equal("https://fixed.example.com/callback"))
	// Match same host.
	g.Expect(client.redirectURIs[3]).To(Equal(requestedRedirect))
	// Naked wildcard.
	g.Expect(client.redirectURIs[4]).To(Equal(requestedRedirect))
}

// TestClientInjectWildcardPathPattern tests wildcard with path components.
func TestClientInjectWildcardPathPattern(t *testing.T) {
	g := NewGomegaWithT(t)

	issuer := "https://hub.example.com/auth/oidc"
	requestedRedirect := "https://hub.example.com/auth/callback"
	req := httptest.NewRequest(
		"GET",
		issuer+"/authorize?redirect_uri="+requestedRedirect,
		nil,
	)

	client := &Client{
		id:              "path-wildcard",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs: []string{
			"https://hub.example.com/**/callback", // Doublestar matches multi-level
			"https://hub.example.com/*/callback",  // Single * matches single level
		},
		request: req,
	}

	client.Inject()

	// Doublestar pattern matches requested redirect
	g.Expect(client.redirectURIs[0]).To(Equal(requestedRedirect))
	// Single wildcard also matches (auth is single path segment)
	g.Expect(client.redirectURIs[1]).To(Equal(requestedRedirect))
}

// TestClientInjectNativeApplication tests injection with native application type.
func TestClientInjectNativeApplication(t *testing.T) {
	g := NewGomegaWithT(t)

	m := &IdpClient{
		ClientId:        "native-client",
		ApplicationType: "native",
		Grants:          []string{"authorization_code", "urn:ietf:params:oauth:grant-type:device_code"},
		RedirectURIs:    []string{"http://localhost:*/callback", "urn:ietf:wg:oauth:2.0:oob"},
		Scopes:          []string{"openid"},
	}

	req := httptest.NewRequest("GET", "http://localhost:8080/oidc", nil)
	client := &Client{}
	client.With(m, req)

	g.Expect(client.applicationType).To(Equal(op.ApplicationTypeNative))
	g.Expect(client.grantTypes).To(ContainElement("urn:ietf:params:oauth:grant-type:device_code"))

	// Now inject with updated request
	issuer := "http://localhost:8080/oidc"
	requestedRedirect := "http://localhost:8080/callback"
	req2 := httptest.NewRequest(
		"GET",
		issuer+"/authorize?redirect_uri="+requestedRedirect,
		nil,
	)

	client.request = req2
	client.Inject()

	// Wildcard matches requested redirect
	g.Expect(client.redirectURIs[0]).To(Equal(requestedRedirect))
	// Out-of-band redirect has no wildcard - unchanged
	g.Expect(client.redirectURIs[1]).To(Equal("urn:ietf:wg:oauth:2.0:oob"))
}

// TestClientInjectEmptyRedirectURIs tests injection with no redirect URIs.
func TestClientInjectEmptyRedirectURIs(t *testing.T) {
	g := NewGomegaWithT(t)

	req := httptest.NewRequest("GET", "http://hub.example.com/oidc/authorize", nil)
	client := &Client{
		id:              "no-redirects",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs:    []string{},
		request:         req,
	}

	client.Inject()

	g.Expect(client.redirectURIs).To(BeEmpty())
}

// TestClientInjectNoRequestedRedirect tests injection when no redirect_uri query param.
func TestClientInjectNoRequestedRedirect(t *testing.T) {
	g := NewGomegaWithT(t)

	req := httptest.NewRequest("GET", "http://hub.example.com/oidc/authorize", nil)
	client := &Client{
		id:              "no-requested",
		applicationType: op.ApplicationTypeWeb,
		redirectURIs:    []string{"http://*/callback"},
		request:         req,
	}

	client.Inject()

	// No requested redirect means wildcard won't match, pattern preserved
	g.Expect(client.redirectURIs).To(HaveLen(1))
	g.Expect(client.redirectURIs[0]).To(Equal("http://*/callback"))
}

// TestClientInjectComplexWildcardPatterns tests various doublestar patterns.
func TestClientInjectComplexWildcardPatterns(t *testing.T) {
	testCases := []struct {
		name              string
		pattern           string
		issuer            string
		requestedRedirect string
		shouldMatch       bool
	}{
		{
			name:              "wildcard subdomain",
			pattern:           "https://*.example.com/callback",
			issuer:            "https://app.example.com/oidc",
			requestedRedirect: "https://app.example.com/callback",
			shouldMatch:       true,
		},
		{
			name:              "wildcard port",
			pattern:           "http://localhost:*/callback",
			issuer:            "http://localhost:8080/oidc",
			requestedRedirect: "http://localhost:8080/callback",
			shouldMatch:       true,
		},
		{
			name:              "wildcard path segment",
			pattern:           "https://hub.io/*/callback",
			issuer:            "https://hub.io/auth/oidc",
			requestedRedirect: "https://hub.io/auth/callback",
			shouldMatch:       true,
		},
		{
			name:              "doublestar path",
			pattern:           "https://hub.io/**/callback",
			issuer:            "https://hub.io/auth/v1/oidc",
			requestedRedirect: "https://hub.io/auth/v1/callback",
			shouldMatch:       true,
		},
		{
			name:              "mismatch scheme",
			pattern:           "http://hub.example.com/callback",
			issuer:            "https://hub.example.com/oidc",
			requestedRedirect: "https://hub.example.com/callback",
			shouldMatch:       false,
		},
		{
			name:              "mismatch host",
			pattern:           "https://app.example.com/callback",
			issuer:            "https://hub.example.org/oidc",
			requestedRedirect: "https://hub.example.org/callback",
			shouldMatch:       false,
		},
		{
			name:              "wildcard entire host",
			pattern:           "https://*/callback",
			issuer:            "https://anything.com/oidc",
			requestedRedirect: "https://anything.com/callback",
			shouldMatch:       true,
		},
		{
			name:              "template with wildcard",
			pattern:           "https://${issuer.host}{,*}/*/callback",
			issuer:            "https://hub.example.com:8080/oidc",
			requestedRedirect: "https://hub.example.com:8080/auth/callback",
			shouldMatch:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			req := httptest.NewRequest(
				"GET",
				tc.issuer+"/authorize?redirect_uri="+tc.requestedRedirect,
				nil,
			)

			client := &Client{
				id:              "pattern-test",
				applicationType: op.ApplicationTypeWeb,
				redirectURIs:    []string{tc.pattern},
				request:         req,
			}

			client.Inject()

			if tc.shouldMatch {
				g.Expect(client.redirectURIs[0]).To(Equal(tc.requestedRedirect), "pattern should match and be replaced")
			} else {
				g.Expect(client.redirectURIs[0]).To(Equal(tc.pattern), "pattern should not match, remain unchanged")
			}
		})
	}
}

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB() (db *gorm.DB, err error) {
	db, err = database.OpenTest()
	if err != nil {
		return
	}

	// Auto-migrate test models
	err = db.AutoMigrate(
		&IdpClient{},
		&User{},
		&model.Task{},
		&model.Bucket{},
		&Role{},
		&Token{},
		&Grant{},
		&RsaKey{},
		&Identity{},
	)
	return
}

// mockRelyingParty implements rp.RelyingParty for testing.
type mockRelyingParty struct {
	endSessionEndpoint string
}

func (m *mockRelyingParty) OAuthConfig() *oauth2.Config              { return &oauth2.Config{} }
func (m *mockRelyingParty) Issuer() string                           { return "" }
func (m *mockRelyingParty) IsPKCE() bool                             { return false }
func (m *mockRelyingParty) CookieHandler() *httphelper.CookieHandler { return nil }
func (m *mockRelyingParty) HttpClient() *http.Client                 { return nil }
func (m *mockRelyingParty) IsOAuth2Only() bool                       { return false }
func (m *mockRelyingParty) Signer() jose.Signer                      { return nil }
func (m *mockRelyingParty) GetEndSessionEndpoint() string            { return m.endSessionEndpoint }
func (m *mockRelyingParty) GetRevokeEndpoint() string                { return "" }
func (m *mockRelyingParty) UserinfoEndpoint() string                 { return "" }
func (m *mockRelyingParty) GetDeviceAuthorizationEndpoint() string   { return "" }
func (m *mockRelyingParty) IDTokenVerifier() *rp.IDTokenVerifier     { return nil }
func (m *mockRelyingParty) ErrorHandler() func(http.ResponseWriter, *http.Request, string, string, string) {
	return nil
}
func (m *mockRelyingParty) Logger(context.Context) (*slog.Logger, bool) { return nil, false }

// TestEndSessionURL tests building the upstream IdP end_session URL.
func TestEndSessionURL(t *testing.T) {
	g := NewGomegaWithT(t)

	savedIdp := federated.Idp
	defer func() { federated.Idp = savedIdp }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/logout",
		},
	}

	logoutURL, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(logoutURL).NotTo(BeEmpty())
	u, err := url.Parse(logoutURL)
	g.Expect(err).To(BeNil())
	g.Expect(u.Scheme).To(Equal("https"))
	g.Expect(u.Host).To(Equal("keycloak.example.com"))
	g.Expect(u.Query().Get("client_id")).To(Equal("hub-client"))
	g.Expect(u.Query().Get("post_logout_redirect_uri")).To(Equal("https://app.example.com/"))
}

// TestEndSessionURLNoRedirect tests building the URL without a post_logout_redirect_uri.
func TestEndSessionURLNoRedirect(t *testing.T) {
	g := NewGomegaWithT(t)

	savedIdp := federated.Idp
	defer func() { federated.Idp = savedIdp }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/logout",
		},
	}

	logoutURL, err := h.EndSessionURL("")
	g.Expect(err).To(BeNil())
	g.Expect(logoutURL).NotTo(BeEmpty())
	u, err := url.Parse(logoutURL)
	g.Expect(err).To(BeNil())
	g.Expect(u.Query().Get("client_id")).To(Equal("hub-client"))
	g.Expect(u.Query().Get("post_logout_redirect_uri")).To(Equal(""))
}

// TestEndSessionURLExistingQuery tests that existing query parameters
// on the end_session_endpoint are preserved.
func TestEndSessionURLExistingQuery(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/logout?foo=bar",
		},
	}

	logoutURL, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(logoutURL).NotTo(BeEmpty())
	u, err := url.Parse(logoutURL)
	g.Expect(err).To(BeNil())
	g.Expect(u.Query().Get("foo")).To(Equal("bar"))
	g.Expect(u.Query().Get("client_id")).To(Equal("hub-client"))
	g.Expect(u.Query().Get("post_logout_redirect_uri")).To(Equal("https://app.example.com/"))
}

// TestEndSessionURLDisabled tests that EndSessionURL returns empty when IdP is disabled.
func TestEndSessionURLDisabled(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled: false,
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/logout",
		},
	}

	logoutURL, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(logoutURL).To(BeEmpty())
}

// TestEndSessionURLNoEndpoint tests that EndSessionURL returns empty
// when the IdP has no end_session_endpoint.
func TestEndSessionURLNoEndpoint(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "",
		},
	}

	logoutURL, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(logoutURL).To(BeEmpty())
}

// TestEndSessionURLNoClient tests that EndSessionURL returns an error
// when the RP client cannot be initialized.
func TestEndSessionURLNoClient(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{}

	logoutURL, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).NotTo(BeNil())
	g.Expect(logoutURL).To(BeEmpty())
}

// TestCreateAccessToken_UpdatesExistingToken tests that refreshing a token
// updates the existing token record instead of creating a new one.
func TestCreateAccessToken_UpdatesExistingToken(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := database.OpenTest()
	g.Expect(err).To(BeNil())
	db.AutoMigrate(
		&IdpClient{},
		&User{},
		&Task{},
		&model.Bucket{},
		&Role{},
		&Token{},
		&Grant{},
		&RsaKey{},
		&Identity{})

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	user := &User{Login: "testuser"}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())
	provider.cache.UserSaved(user)

	subject := &Subject{}
	scopes, err := provider.cache.FindScopes(user.Subject)
	g.Expect(err).To(BeNil())
	subject.WithUser(user, scopes)

	grantId := provider.storage.genId()
	g.Expect(err).To(BeNil())

	refreshReq := &RefreshRequest{
		grantId:  grantId,
		clientId: "test-client",
		subject:  subject.Key,
		scopes:   []string{"openid", "profile"},
		issued:   time.Now(),
	}

	tokenId1, expiration1, err := provider.storage.CreateAccessToken(context.Background(), refreshReq)
	g.Expect(err).To(BeNil())
	g.Expect(tokenId1).To(Equal(grantId))

	var tokens1 []Token
	err = db.Find(&tokens1, "authId = ?", tokenId1).Error
	g.Expect(err).To(BeNil())
	g.Expect(tokens1).To(HaveLen(1))

	firstToken := tokens1[0]
	firstScopes := firstToken.Scopes

	time.Sleep(100 * time.Millisecond)

	refreshReq.scopes = []string{"openid", "profile", "email"}
	tokenId2, expiration2, err := provider.storage.CreateAccessToken(context.Background(), refreshReq)
	g.Expect(err).To(BeNil())
	g.Expect(tokenId2).To(Equal(grantId))
	g.Expect(tokenId2).To(Equal(tokenId1))

	var tokens2 []Token
	err = db.Find(&tokens2, "authId = ?", tokenId1).Error
	g.Expect(err).To(BeNil())
	g.Expect(tokens2).To(HaveLen(1))

	g.Expect(expiration2).To(BeTemporally(">", expiration1))

	g.Expect(tokens2[0].Scopes).To(Equal([]string{"openid", "profile", "email"}))
	g.Expect(tokens2[0].Scopes).NotTo(Equal(firstScopes))

	g.Expect(tokens2[0].Subject).To(Equal(firstToken.Subject))
	g.Expect(tokens2[0].Issued).To(Equal(firstToken.Issued))
	g.Expect(tokens2[0].ID).To(Equal(firstToken.ID))
}

// TestCreateAccessToken_CascadeDeleteOnGrantDeletion tests that deleting
// a grant CASCADE deletes its associated access token.
func TestCreateAccessToken_CascadeDeleteOnGrantDeletion(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := database.OpenTest()
	g.Expect(err).To(BeNil())
	db.AutoMigrate(
		&IdpClient{},
		&User{},
		&Task{},
		&model.Bucket{},
		&Role{},
		&Token{},
		&Grant{},
		&RsaKey{},
		&Identity{})

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	user := &User{Login: "testuser"}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())
	provider.cache.UserSaved(user)

	subject := &Subject{}
	scopes, err := provider.cache.FindScopes(user.Subject)
	g.Expect(err).To(BeNil())
	subject.WithUser(user, scopes)

	grantId := provider.storage.genId()
	grant := &Grant{
		Kind:    KindAuthCode,
		AuthId:  grantId,
		Subject: subject.Key,
		Scopes:  []string{"openid profile"},
		Issued:  time.Now(),
	}
	err = db.Create(grant).Error

	refreshReq := &RefreshRequest{
		grantId:  grantId,
		clientId: "test-client",
		subject:  subject.Key,
		scopes:   []string{"openid", "profile"},
		issued:   time.Now(),
	}

	tokenId, _, err := provider.storage.CreateAccessToken(context.Background(), refreshReq)
	g.Expect(err).To(BeNil())

	var tokens []Token
	err = db.Find(&tokens, "authId = ?", tokenId).Error
	g.Expect(err).To(BeNil())
	g.Expect(tokens).To(HaveLen(1))

	err = db.Delete(grant).Error
	g.Expect(err).To(BeNil())

	var tokensAfter []Token
	err = db.Find(&tokensAfter, "authId = ?", tokenId).Error
	g.Expect(err).To(BeNil())
	g.Expect(tokensAfter).To(BeEmpty())
}

// TestAuthRequest_CreatesGrantAndLinksToken tests that AuthRequest creates
// a grant first, then creates a token linked to that grant.
func TestAuthRequest_CreatesGrantAndLinksToken(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := database.OpenTest()
	g.Expect(err).To(BeNil())
	db.AutoMigrate(
		&IdpClient{},
		&User{},
		&Task{},
		&model.Bucket{},
		&Role{},
		&Token{},
		&Grant{},
		&RsaKey{},
		&Identity{})

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	user := &User{Login: "testuser"}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())
	provider.cache.UserSaved(user)

	subject := &Subject{}
	scopes, err := provider.cache.FindScopes(user.Subject)
	g.Expect(err).To(BeNil())
	subject.WithUser(user, scopes)

	// Create AuthRequest
	authReq := &AuthRequest{
		requestId: provider.storage.genId(),
		subject:   subject.Key,
		AuthRequest: &oidc.AuthRequest{
			ClientID: "test-client",
			Scopes:   []string{"openid", "profile"},
		},
		issued: time.Now(),
	}

	// CreateAccessToken with AuthRequest should create grant first
	tokenId, _, err := provider.storage.CreateAccessToken(context.Background(), authReq)
	g.Expect(err).To(BeNil())
	g.Expect(tokenId).To(Equal(authReq.GetID()))

	// Verify grant was created with matching authId
	var grant Grant
	err = db.First(&grant, "authId = ?", authReq.GetID()).Error
	g.Expect(err).To(BeNil())
	g.Expect(grant.AuthId).To(Equal(authReq.GetID()))
	g.Expect(grant.Subject).To(Equal(subject.Key))

	// Verify token was created and linked to grant
	var token Token
	err = db.First(&token, "authId = ?", authReq.GetID()).Error
	g.Expect(err).To(BeNil())
	g.Expect(token.AuthId).To(Equal(authReq.GetID()))
	g.Expect(token.GrantID).NotTo(BeNil())
	g.Expect(*token.GrantID).To(Equal(grant.ID))
}

// TestClientRequest_NoGrantCreated tests that ClientRequest creates tokens
// without creating grants, and each request gets a new token.
func TestClientRequest_NoGrantCreated(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := database.OpenTest()
	g.Expect(err).To(BeNil())
	db.AutoMigrate(
		&IdpClient{},
		&User{},
		&Task{},
		&model.Bucket{},
		&Role{},
		&Token{},
		&Grant{},
		&RsaKey{},
		&Identity{})

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create IdP client in database and cache
	client := &IdpClient{
		Subject:         uuid.New().String(),
		ClientId:        "test-client-id",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	err = db.Create(client).Error
	g.Expect(err).To(BeNil())
	provider.cache.ClientSaved(client)

	// Create two client requests with different authIds
	clientReq1 := &ClientRequest{
		authId:   provider.storage.genId(),
		clientId: "test-client-id",
		subject:  client.Subject,
		scopes:   []string{"openid"},
		issued:   time.Now(),
	}

	clientReq2 := &ClientRequest{
		authId:   provider.storage.genId(),
		clientId: "test-client-id",
		subject:  client.Subject,
		scopes:   []string{"openid"},
		issued:   time.Now(),
	}

	// Create first token
	tokenId1, _, err := provider.storage.CreateAccessToken(context.Background(), clientReq1)
	g.Expect(err).To(BeNil())
	g.Expect(tokenId1).To(Equal(clientReq1.authId))

	// Verify no grant was created
	var grants []Grant
	err = db.Find(&grants, "authId = ?", clientReq1.authId).Error
	g.Expect(err).To(BeNil())
	g.Expect(grants).To(BeEmpty())

	// Verify token was created with nil GrantID
	var token1 Token
	err = db.First(&token1, "authId = ?", clientReq1.authId).Error
	g.Expect(err).To(BeNil())
	g.Expect(token1.GrantID).To(BeNil())

	// Create second token (different authId)
	tokenId2, _, err := provider.storage.CreateAccessToken(context.Background(), clientReq2)
	g.Expect(err).To(BeNil())
	g.Expect(tokenId2).To(Equal(clientReq2.authId))
	g.Expect(tokenId2).NotTo(Equal(tokenId1))

	// Verify second token was created (not upserted)
	var tokens []Token
	err = db.Find(&tokens).Error
	g.Expect(err).To(BeNil())
	g.Expect(tokens).To(HaveLen(2))
}

// TestCreateAccessAndRefreshTokens_FullFlow tests the complete flow:
// AuthRequest creates grant and token, then refresh token updates the grant.
func TestCreateAccessAndRefreshTokens_FullFlow(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := database.OpenTest()
	g.Expect(err).To(BeNil())
	db.AutoMigrate(
		&IdpClient{},
		&User{},
		&Task{},
		&model.Bucket{},
		&Role{},
		&Token{},
		&Grant{},
		&RsaKey{},
		&Identity{})

	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	user := &User{Login: "testuser"}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())
	provider.cache.UserSaved(user)

	subject := &Subject{}
	scopes, err := provider.cache.FindScopes(user.Subject)
	g.Expect(err).To(BeNil())
	subject.WithUser(user, scopes)

	// Create AuthRequest
	requestId := provider.storage.genId()
	authReq := &AuthRequest{
		requestId: requestId,
		subject:   subject.Key,
		AuthRequest: &oidc.AuthRequest{
			ClientID: "test-client",
			Scopes:   []string{"openid", "profile", "offline_access"},
		},
		issued: time.Now(),
	}

	// Store authReq so createRefreshToken can find it
	provider.storage.mutex.Lock()
	provider.storage.authReqById[requestId] = authReq
	provider.storage.mutex.Unlock()

	// Call CreateAccessAndRefreshTokens
	accessTokenId, refreshToken, _, err := provider.storage.CreateAccessAndRefreshTokens(
		context.Background(),
		authReq,
		"")
	g.Expect(err).To(BeNil())
	g.Expect(accessTokenId).To(Equal(requestId))
	g.Expect(refreshToken).NotTo(BeEmpty())

	// Verify grant was created
	var grant Grant
	err = db.First(&grant, "authId = ?", requestId).Error
	g.Expect(err).To(BeNil())
	g.Expect(grant.AuthId).To(Equal(requestId))

	// Verify grant has refresh token hash
	g.Expect(grant.RefreshToken).NotTo(BeEmpty())

	// Verify access token was created and linked to grant
	var token Token
	err = db.First(&token, "authId = ?", requestId).Error
	g.Expect(err).To(BeNil())
	g.Expect(token.AuthId).To(Equal(requestId))
	g.Expect(token.GrantID).NotTo(BeNil())
	g.Expect(*token.GrantID).To(Equal(grant.ID))
}

// TestUserFilter tests placeholder replacement in user filters.
func TestUserFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	ds := &LDAP{}

	// Test ${uid} replacement
	ds.UserFilter = "(uid=${uid})"
	filter := ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(uid=jsmith)"))

	// Test ${login} replacement
	ds.UserFilter = "(sAMAccountName=${login})"
	filter = ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(sAMAccountName=jsmith)"))

	// Test both ${uid} and ${login} in same filter
	ds.UserFilter = "(|(uid=${uid})(mail=${login}))"
	filter = ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(|(uid=jsmith)(mail=jsmith))"))

	// Test LDAP escaping of special characters
	ds.UserFilter = "(uid=${uid})"
	filter = ds.userFilter("j*smith")
	g.Expect(filter).To(Equal("(uid=j\\2asmith)"))

	// Test LDAP escaping of parentheses
	ds.UserFilter = "(uid=${login})"
	filter = ds.userFilter("j(smith)")
	g.Expect(filter).To(Equal("(uid=j\\28smith\\29)"))

	// Test Active Directory default
	ds.Kind = "AD"
	ds.UserFilter = ""
	filter = ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(sAMAccountName=jsmith)"))

	// Test ACTIVEDIRECTORY kind
	ds.Kind = "ACTIVEDIRECTORY"
	ds.UserFilter = ""
	filter = ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(sAMAccountName=jsmith)"))

	// Test standard LDAP default
	ds.Kind = "LDAP"
	ds.UserFilter = ""
	filter = ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(uid=jsmith)"))

	// Test unknown kind defaults to LDAP behavior
	ds.Kind = "UNKNOWN"
	ds.UserFilter = ""
	filter = ds.userFilter("jsmith")
	g.Expect(filter).To(Equal("(uid=jsmith)"))

	// Test custom filter with multiple placeholders and escaping
	ds.UserFilter = "(&(uid=${uid})(mail=${login}@example.com))"
	filter = ds.userFilter("test.user")
	g.Expect(filter).To(Equal("(&(uid=test.user)(mail=test.user@example.com))"))
}

// TestGroupFilter tests placeholder replacement in group filters.
func TestGroupFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	ds := &LDAP{}

	// Create mock user entry
	user := &ldap.Entry{
		DN: "uid=jsmith,ou=people,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{Name: "cn", Values: []string{"John Smith"}},
			{Name: "uid", Values: []string{"jsmith"}},
		},
	}

	// Test ${dn} replacement
	ds.GroupFilter = "(member=${dn})"
	filter := ds.groupFilter(user)
	g.Expect(filter).To(Equal("(member=uid=jsmith,ou=people,dc=example,dc=com)"))

	// Test ${cn} replacement
	ds.GroupFilter = "(owner=${cn})"
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(owner=John Smith)"))

	// Test ${uid} replacement
	ds.GroupFilter = "(memberUid=${uid})"
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(memberUid=jsmith)"))

	// Test multiple placeholders
	ds.GroupFilter = "(&(member=${dn})(owner=${cn}))"
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(&(member=uid=jsmith,ou=people,dc=example,dc=com)(owner=John Smith))"))

	// Test all three placeholders
	ds.GroupFilter = "(&(member=${dn})(owner=${cn})(memberUid=${uid}))"
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(&(member=uid=jsmith,ou=people,dc=example,dc=com)(owner=John Smith)(memberUid=jsmith))"))

	// Test LDAP escaping in DN
	userWithSpecialDN := &ldap.Entry{
		DN: "uid=j*smith,ou=people,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{Name: "cn", Values: []string{"J*Smith"}},
			{Name: "uid", Values: []string{"j*smith"}},
		},
	}
	ds.GroupFilter = "(member=${dn})"
	filter = ds.groupFilter(userWithSpecialDN)
	g.Expect(filter).To(Equal("(member=uid=j\\2asmith,ou=people,dc=example,dc=com)"))

	// Test LDAP escaping in CN
	ds.GroupFilter = "(owner=${cn})"
	filter = ds.groupFilter(userWithSpecialDN)
	g.Expect(filter).To(Equal("(owner=J\\2aSmith)"))

	// Test Active Directory default
	ds.Kind = "AD"
	ds.GroupFilter = ""
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(&(objectClass=group)(member=uid=jsmith,ou=people,dc=example,dc=com))"))

	// Test ACTIVEDIRECTORY kind
	ds.Kind = "ACTIVEDIRECTORY"
	ds.GroupFilter = ""
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(&(objectClass=group)(member=uid=jsmith,ou=people,dc=example,dc=com))"))

	// Test standard LDAP default
	ds.Kind = "LDAP"
	ds.GroupFilter = ""
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(&(objectClass=*)(member=uid=jsmith,ou=people,dc=example,dc=com))"))

	// Test unknown kind defaults to LDAP behavior
	ds.Kind = "UNKNOWN"
	ds.GroupFilter = ""
	filter = ds.groupFilter(user)
	g.Expect(filter).To(Equal("(&(objectClass=*)(member=uid=jsmith,ou=people,dc=example,dc=com))"))

	// Test empty attribute values
	userNoAttrs := &ldap.Entry{
		DN:         "uid=noattrs,ou=people,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{},
	}
	ds.GroupFilter = "(&(member=${dn})(memberUid=${uid}))"
	filter = ds.groupFilter(userNoAttrs)
	g.Expect(filter).To(Equal("(&(member=uid=noattrs,ou=people,dc=example,dc=com)(memberUid=))"))
}

// TestScopeExpand tests Scope.Expand() with wildcard patterns.
func TestScopeExpand(t *testing.T) {
	g := NewGomegaWithT(t)

	// Register test resources in Tenant
	Domain.Register("applications")
	Domain.Register("tags")
	Domain.Register("identities")

	// Test 1: Wildcard resource and method (*:*)
	scope := Scope{Resource: "*", Method: "*"}
	expanded := scope.Expand()
	g.Expect(len(expanded)).To(BeNumerically(">", 0))

	// Verify all verbs are present for at least one resource
	foundGet := false
	foundPost := false
	foundDelete := false
	foundDecrypt := false
	for _, s := range expanded {
		if s.Method == "get" {
			foundGet = true
		}
		if s.Method == "post" {
			foundPost = true
		}
		if s.Method == "delete" {
			foundDelete = true
		}
		if s.Method == "decrypt" {
			foundDecrypt = true
		}
	}
	g.Expect(foundGet).To(BeTrue())
	g.Expect(foundPost).To(BeTrue())
	g.Expect(foundDelete).To(BeTrue())
	g.Expect(foundDecrypt).To(BeTrue())

	// Test 2: Wildcard method (applications:*)
	scope = Scope{Resource: "applications", Method: "*"}
	expanded = scope.Expand()
	g.Expect(expanded).To(HaveLen(6)) // 6 verbs

	// All should be for applications
	for _, s := range expanded {
		g.Expect(s.Resource).To(Equal("applications"))
	}

	// Verify all verbs present
	methods := make(map[string]bool)
	for _, s := range expanded {
		methods[s.Method] = true
	}
	g.Expect(methods).To(HaveKey("decrypt"))
	g.Expect(methods).To(HaveKey("delete"))
	g.Expect(methods).To(HaveKey("get"))
	g.Expect(methods).To(HaveKey("patch"))
	g.Expect(methods).To(HaveKey("post"))
	g.Expect(methods).To(HaveKey("put"))

	// Test 3: Wildcard resource (*:get)
	scope = Scope{Resource: "*", Method: "get"}
	expanded = scope.Expand()
	g.Expect(len(expanded)).To(BeNumerically(">", 0))

	// All should have method "get"
	for _, s := range expanded {
		g.Expect(s.Method).To(Equal("get"))
	}

	// Should have multiple resources
	resources := make(map[string]bool)
	for _, s := range expanded {
		resources[s.Resource] = true
	}
	g.Expect(len(resources)).To(BeNumerically(">", 1))

	// Test 4: Exact scope (no wildcards) - applications:get
	scope = Scope{Resource: "applications", Method: "get"}
	expanded = scope.Expand()
	g.Expect(expanded).To(HaveLen(1))
	g.Expect(expanded[0].Resource).To(Equal("applications"))
	g.Expect(expanded[0].Method).To(Equal("get"))

	// Test 5: Another exact scope - tags:delete
	scope = Scope{Resource: "tags", Method: "delete"}
	expanded = scope.Expand()
	g.Expect(expanded).To(HaveLen(1))
	g.Expect(expanded[0].Resource).To(Equal("tags"))
	g.Expect(expanded[0].Method).To(Equal("delete"))
}

// TestExpandScopes tests the ExpandScopes helper function.
func TestExpandScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	// Register test resources in Tenant
	Domain.Register("applications")
	Domain.Register("tags")

	// Test 1: Expand single wildcard scope
	scopes := ExpandScopes("applications:*")
	g.Expect(scopes).To(HaveLen(6)) // 6 verbs
	g.Expect(scopes).To(ContainElement("applications:decrypt"))
	g.Expect(scopes).To(ContainElement("applications:delete"))
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("applications:patch"))
	g.Expect(scopes).To(ContainElement("applications:post"))
	g.Expect(scopes).To(ContainElement("applications:put"))

	// Test 2: Expand multiple scopes with wildcards
	scopes = ExpandScopes("applications:*", "tags:get")
	g.Expect(len(scopes)).To(Equal(7)) // 6 from applications:* + 1 from tags:get
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("tags:get"))

	// Test 3: Expand *:* (all resources, all methods)
	scopes = ExpandScopes("*:*")
	g.Expect(len(scopes)).To(BeNumerically(">", 10)) // Many resources × 6 verbs

	// Test 4: Mix of wildcard and exact scopes
	scopes = ExpandScopes("applications:get", "tags:*", "identities:post")
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("identities:post"))
	g.Expect(scopes).To(ContainElement("tags:delete"))
	g.Expect(scopes).To(ContainElement("tags:get"))

	// Test 5: Empty input
	scopes = ExpandScopes()
	g.Expect(scopes).To(BeEmpty())

	// Test 6: Exact scopes (no expansion)
	scopes = ExpandScopes("applications:get", "tags:post")
	g.Expect(scopes).To(HaveLen(2))
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("tags:post"))
}

// TestWildcardScopeMatching tests that Match() handles wildcards correctly.
func TestWildcardScopeMatching(t *testing.T) {
	g := NewGomegaWithT(t)

	// Test 1: *:* matches everything
	scope := Scope{Resource: "*", Method: "*"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("tags", "POST")).To(BeTrue())
	g.Expect(scope.Match("anything", "DELETE")).To(BeTrue())
	g.Expect(scope.Match("admin", "decrypt")).To(BeTrue())

	// Test 2: resource:* matches any method for that resource
	scope = Scope{Resource: "applications", Method: "*"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeTrue())
	g.Expect(scope.Match("applications", "DELETE")).To(BeTrue())
	g.Expect(scope.Match("applications", "decrypt")).To(BeTrue())
	g.Expect(scope.Match("tags", "GET")).To(BeFalse())
	g.Expect(scope.Match("identities", "POST")).To(BeFalse())

	// Test 3: *:method matches that method for any resource
	scope = Scope{Resource: "*", Method: "GET"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("tags", "GET")).To(BeTrue())
	g.Expect(scope.Match("anything", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeFalse())
	g.Expect(scope.Match("tags", "DELETE")).To(BeFalse())

	// Test 4: Exact match (no wildcards)
	scope = Scope{Resource: "applications", Method: "GET"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("applications", "POST")).To(BeFalse())
	g.Expect(scope.Match("tags", "GET")).To(BeFalse())
	g.Expect(scope.Match("tags", "POST")).To(BeFalse())

	// Test 5: Case insensitive matching
	scope = Scope{Resource: "Applications", Method: "get"}
	g.Expect(scope.Match("applications", "GET")).To(BeTrue())
	g.Expect(scope.Match("APPLICATIONS", "get")).To(BeTrue())

	scope = Scope{Resource: "*", Method: "Get"}
	g.Expect(scope.Match("tags", "GET")).To(BeTrue())
	g.Expect(scope.Match("Tags", "get")).To(BeTrue())

	// Test 6: Special verb "decrypt"
	scope = Scope{Resource: "admin", Method: "*"}
	g.Expect(scope.Match("admin", "decrypt")).To(BeTrue())

	scope = Scope{Resource: "*", Method: "decrypt"}
	g.Expect(scope.Match("applications", "decrypt")).To(BeTrue())
	g.Expect(scope.Match("admin", "decrypt")).To(BeTrue())
}

// TestScopeString tests Scope.String() method.
func TestScopeString(t *testing.T) {
	g := NewGomegaWithT(t)

	scope := Scope{Resource: "applications", Method: "get"}
	g.Expect(scope.String()).To(Equal("applications:get"))

	scope = Scope{Resource: "*", Method: "*"}
	g.Expect(scope.String()).To(Equal("*:*"))

	scope = Scope{Resource: "tags", Method: "*"}
	g.Expect(scope.String()).To(Equal("tags:*"))

	scope = Scope{Resource: "*", Method: "delete"}
	g.Expect(scope.String()).To(Equal("*:delete"))
}

// TestAdminWildcardScopes tests that admin role gets all scopes via wildcards.
func TestAdminWildcardScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Register resources
	// Resources are now handled by Expand() internally for tests
	// Resources are now handled by Expand() internally for tests

	// Create admin role with wildcard scopes
	adminRole := &model.Role{
		Name:   "admin",
		Scopes: []string{"admin:*", "*:*"},
	}
	err = db.Create(adminRole).Error
	g.Expect(err).To(BeNil())

	// Create admin user
	adminUser := &model.User{
		Subject:  "admin-wildcard-user",
		Login:    "adminwildcard",
		Password: secret.HashPassword("password"),
		Email:    "adminwildcard@example.com",
	}
	err = db.Create(adminUser).Error
	g.Expect(err).To(BeNil())

	err = db.Model(adminUser).Association("Roles").Append(adminRole)
	g.Expect(err).To(BeNil())

	// Create provider
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create token
	token, err := provider.NewToken(adminUser.Subject, 24*time.Hour)
	g.Expect(err).To(BeNil())

	// Authenticate
	request := newTestRequest()
	request.With("Bearer " + token.Secret)
	jwToken, err := provider.Authenticate(request)
	g.Expect(err).To(BeNil())

	// Get scopes from token (stored in wildcard form)
	scopes := provider.Scopes(jwToken)
	scopeStrings := make([]string, len(scopes))
	for i, s := range scopes {
		scopeStrings[i] = s.String()
	}

	// Verify token contains wildcard scopes (not expanded)
	g.Expect(scopeStrings).To(ContainElement("admin:*"))
	g.Expect(scopeStrings).To(ContainElement("*:*"))

	// Verify wildcard scopes match specific operations
	// admin:* should match admin with any method
	for _, scope := range scopes {
		if scope.Resource == "admin" && scope.Method == "*" {
			g.Expect(scope.Match("admin", "decrypt")).To(BeTrue())
			g.Expect(scope.Match("admin", "get")).To(BeTrue())
			g.Expect(scope.Match("admin", "post")).To(BeTrue())
			g.Expect(scope.Match("admin", "delete")).To(BeTrue())
			g.Expect(scope.Match("applications", "get")).To(BeFalse())
		}
	}

	// *:* should match any resource with any method
	for _, scope := range scopes {
		if scope.Resource == "*" && scope.Method == "*" {
			g.Expect(scope.Match("applications", "get")).To(BeTrue())
			g.Expect(scope.Match("applications", "decrypt")).To(BeTrue())
			g.Expect(scope.Match("tags", "delete")).To(BeTrue())
			g.Expect(scope.Match("admin", "decrypt")).To(BeTrue())
			g.Expect(scope.Match("anything", "anymethod")).To(BeTrue())
		}
	}
}

// TestExternalIdpWildcardExpansion tests wildcard expansion from external IdP tokens.
func TestExternalIdpWildcardExpansion(t *testing.T) {
	g := NewGomegaWithT(t)

	// Register test resources in Tenant
	Domain.Register("applications")
	Domain.Register("tags")

	// Simulate external IdP scopes with wildcards
	idpScopes := []string{
		"applications:*",
		"tags:get",
		"*:delete",
	}

	// Expand using ExpandScopes (same function used in FedIdpLogin.extractScopes)
	expanded := ExpandScopes(idpScopes...)

	// Verify applications:* expanded to all verbs
	g.Expect(expanded).To(ContainElement("applications:decrypt"))
	g.Expect(expanded).To(ContainElement("applications:delete"))
	g.Expect(expanded).To(ContainElement("applications:get"))
	g.Expect(expanded).To(ContainElement("applications:patch"))
	g.Expect(expanded).To(ContainElement("applications:post"))
	g.Expect(expanded).To(ContainElement("applications:put"))

	// Verify tags:get remains unchanged
	g.Expect(expanded).To(ContainElement("tags:get"))

	// Verify *:delete expanded to all resources
	g.Expect(expanded).To(ContainElement("applications:delete"))
	g.Expect(expanded).To(ContainElement("tags:delete"))
}

// TestFedIdpRoleScopeExpansion tests role reference expansion in federated IdP scopes.
func TestFedIdpRoleScopeExpansion(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create roles with scopes (scopes are just strings in the Role model)
	adminRole := &model.Role{
		Name:   "admin",
		Scopes: []string{"applications:get", "applications:post"},
	}
	err = db.Create(adminRole).Error
	g.Expect(err).To(BeNil())

	viewerRole := &model.Role{
		Name:   "viewer",
		Scopes: []string{"tags:get"},
	}
	err = db.Create(viewerRole).Error
	g.Expect(err).To(BeNil())

	// Create provider and cache
	provider, err := NewBuiltin(db)
	g.Expect(err).To(BeNil())

	// Create FedIdpHandler and FedIdpLogin
	handler := &FedIdpHandler{
		cache: provider.cache,
	}

	login := &FedIdpLogin{
		handler:           handler,
		accessTokenClaims: make(map[string]any),
	}

	// Test 1: Expand +role.admin to admin role's scopes
	login.accessTokenClaims[ClaimScope] = "+role.admin applications:delete"
	scopes := login.extractScopes()
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("applications:post"))
	g.Expect(scopes).To(ContainElement("applications:delete"))
	g.Expect(scopes).NotTo(ContainElement("+role.admin"))

	// Test 2: Mix of role reference and regular scopes
	login.accessTokenClaims[ClaimScope] = "+role.viewer applications:put"
	scopes = login.extractScopes()
	g.Expect(scopes).To(ContainElement("tags:get"))
	g.Expect(scopes).To(ContainElement("applications:put"))
	g.Expect(scopes).NotTo(ContainElement("+role.viewer"))

	// Test 3: Multiple role references
	login.accessTokenClaims[ClaimScope] = "+role.admin +role.viewer"
	scopes = login.extractScopes()
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("applications:post"))
	g.Expect(scopes).To(ContainElement("tags:get"))

	// Test 4: Unknown role - should be logged and dropped
	login.accessTokenClaims[ClaimScope] = "+role.unknown tags:post"
	scopes = login.extractScopes()
	g.Expect(scopes).To(ContainElement("tags:post"))
	g.Expect(scopes).NotTo(ContainElement("+role.unknown"))
	g.Expect(len(scopes)).To(Equal(1))

	// Test 5: No role references - regular scopes unchanged
	login.accessTokenClaims[ClaimScope] = "applications:get tags:post"
	scopes = login.extractScopes()
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("tags:post"))
	g.Expect(len(scopes)).To(Equal(2))

	// Test 6: Empty scopes
	login.accessTokenClaims[ClaimScope] = ""
	scopes = login.extractScopes()
	g.Expect(scopes).To(BeEmpty())

	// Test 7: Role with no scopes
	emptyRole := &model.Role{Name: "empty"}
	err = db.Create(emptyRole).Error
	g.Expect(err).To(BeNil())
	err = provider.cache.Refresh()
	g.Expect(err).To(BeNil())

	login.accessTokenClaims[ClaimScope] = "+role.empty tags:delete"
	scopes = login.extractScopes()
	g.Expect(scopes).To(ContainElement("tags:delete"))
	g.Expect(len(scopes)).To(Equal(1))
}

// TestScopeGenerationWithNounVerb tests that generateScopes populates Noun and Verb.
func TestScopeGenerationWithNounVerb(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create domain and register resources
	domain := NewTenant(db)
	domain.Register("applications")
	domain.Register("tags")

	// Build scopes
	domain.buildScopes()

	// Verify scopes created (3 resources × 6 verbs: admin is registered by default)
	scopes := domain.Scopes()
	g.Expect(len(scopes)).To(Equal(18))

	// Verify each scope can be parsed
	for _, scopeStr := range scopes {
		g.Expect(scopeStr).NotTo(BeEmpty())

		// Verify Scope format is resource:verb
		scope := Scope{}
		scope.With(scopeStr)
		g.Expect(scope.Resource).NotTo(BeEmpty())
		g.Expect(scope.Method).NotTo(BeEmpty())

		// Verify it's in the registered resources
		validResource := scope.Resource == "applications" || scope.Resource == "tags" || scope.Resource == "admin"
		g.Expect(validResource).To(BeTrue())
	}

	// Verify all verbs present for each resource
	resources := []string{"applications", "tags"}
	for _, resource := range resources {
		foundVerbs := make(map[string]bool)
		for _, scopeStr := range scopes {
			scope := Scope{}
			scope.With(scopeStr)
			if scope.Resource == resource {
				foundVerbs[scope.Method] = true
			}
		}
		g.Expect(foundVerbs).To(HaveKey("decrypt"))
		g.Expect(foundVerbs).To(HaveKey("delete"))
		g.Expect(foundVerbs).To(HaveKey("get"))
		g.Expect(foundVerbs).To(HaveKey("patch"))
		g.Expect(foundVerbs).To(HaveKey("post"))
		g.Expect(foundVerbs).To(HaveKey("put"))
	}
}
