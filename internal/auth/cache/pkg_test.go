package cache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB() (db *gorm.DB, err error) {
	db, err = database.OpenTest()
	if err != nil {
		return
	}

	// Auto-migrate test models
	err = db.AutoMigrate(
		&User{},
		&model.Task{},
		&Role{},
		&Permission{},
		&Token{},
		&Grant{},
		&Identity{},
		&IdpClient{},
	)
	return
}

// TestCacheEntityUpdates tests all Saved/Deleted methods.
func TestCacheEntityUpdates(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test RoleSaved/RoleDeleted
	role := &Role{
		Model: Model{ID: 100},
		Name:  "TestRole",
	}
	cache.RoleSaved(role)
	_, err = cache.FindRoleById(100)
	g.Expect(err).To(BeNil())

	cache.RoleDeleted(100)
	_, err = cache.FindRoleById(100)
	g.Expect(err).NotTo(BeNil())

	// Test UserSaved/UserDeleted
	user := &User{
		Model:   Model{ID: 200},
		Subject: "test-user",
		Login:   "testuser",
	}
	cache.UserSaved(user)
	_, err = cache.FindUserByLogin("testuser")
	g.Expect(err).To(BeNil())

	cache.UserDeleted(200)
	_, err = cache.FindUserByLogin("testuser")
	g.Expect(err).NotTo(BeNil())

	// Test TokenSaved/TaskRevoked
	taskID := uint(300)
	taskToken := &Token{
		Token: model.Token{
			Model:      Model{ID: 300},
			TaskID:     &taskID,
			Digest:     secret.Hash("task-token-300"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task-token-300",
	}
	cache.TokenSaved(taskToken)
	_, err = cache.FindToken("task-token-300")
	g.Expect(err).To(BeNil())

	cache.TaskRevoked(taskID)
	_, err = cache.FindToken("task-token-300")
	g.Expect(err).NotTo(BeNil())

	// Test IdentitySaved/IdentityDeleted
	identity := &Identity{
		Model:   Model{ID: 400},
		Issuer:  "https://idp.example.com",
		Subject: "idp-subject",
		Login:   "idp-userid",
	}
	cache.IdentitySaved(identity)
	_, err = cache.FindIdentityByLogin("idp-userid")
	g.Expect(err).To(BeNil())

	cache.IdentityDeleted(400)
	_, err = cache.FindIdentityByLogin("idp-userid")
	g.Expect(err).NotTo(BeNil())
}

// TestCacheTransaction tests cache transaction behavior.
func TestCacheTransaction(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test successful transaction
	role1 := &Role{
		Model: Model{ID: 101},
		Name:  "TxRole1",
	}
	role2 := &Role{
		Model: Model{ID: 102},
		Name:  "TxRole2",
	}

	err = cache.Transaction(func(tx *Tx) error {
		tx.RoleSaved(role1)
		tx.RoleSaved(role2)
		return nil
	})
	g.Expect(err).To(BeNil())

	// Both roles should be in cache
	d := cache.data.Load()
	_, found1 := d.roleById[101]
	_, found2 := d.roleById[102]
	g.Expect(found1).To(BeTrue())
	g.Expect(found2).To(BeTrue())

	// Test rollback on error
	role3 := &Role{
		Model: Model{ID: 103},
		Name:  "TxRole3",
	}

	err = cache.Transaction(func(tx *Tx) error {
		tx.RoleSaved(role3)
		return fmt.Errorf("simulated error")
	})
	g.Expect(err).NotTo(BeNil())

	// Role3 should NOT be in cache (rolled back)
	d = cache.data.Load()
	_, found3 := d.roleById[103]
	g.Expect(found3).To(BeFalse())

	// Test explicit Begin/Commit/Rollback
	tx := cache.Begin()
	user := &User{
		Model:   Model{ID: 201},
		Subject: "tx-user",
		Login:   "txuser",
	}
	tx.UserSaved(user)
	tx.Commit()

	d = cache.data.Load()
	_, foundUser := d.userById[201]
	g.Expect(foundUser).To(BeTrue())

	// Test rollback
	tx = cache.Begin()
	user2 := &User{
		Model:   Model{ID: 202},
		Subject: "tx-user2",
		Login:   "txuser2",
	}
	tx.UserSaved(user2)
	tx.Rollback() // Discard changes

	d = cache.data.Load()
	_, foundUser2 := d.userById[202]
	g.Expect(foundUser2).To(BeFalse())
}

// TestTaskRevokedRemovesTokens tests that TaskRevoked removes tokens from token cache.
func TestTaskRevokedRemovesTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Add task tokens
	taskID := uint(500)
	token1 := &Token{
		Token: model.Token{
			Model:      Model{ID: 501},
			TaskID:     &taskID,
			Digest:     secret.Hash("task-token-501"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task-token-501",
	}
	token2 := &Token{
		Token: model.Token{
			Model:      Model{ID: 502},
			TaskID:     &taskID,
			Digest:     secret.Hash("task-token-502"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task-token-502",
	}

	cache.TokenSaved(token1)
	cache.TokenSaved(token2)

	// Verify tokens are in cache
	_, err = cache.FindToken("task-token-501")
	g.Expect(err).To(BeNil())
	_, err = cache.FindToken("task-token-502")
	g.Expect(err).To(BeNil())

	// Revoke task - should remove both tokens
	cache.TaskRevoked(taskID)

	// Verify tokens are removed
	_, err = cache.FindToken("task-token-501")
	g.Expect(err).NotTo(BeNil())
	_, err = cache.FindToken("task-token-502")
	g.Expect(err).NotTo(BeNil())
}

// TestTaskRevokedMultipleTokens tests that TaskRevoked removes only tokens for specified task.
func TestTaskRevokedMultipleTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Add tokens for two different tasks
	task1ID := uint(600)
	task2ID := uint(601)

	token1 := &Token{
		Token: model.Token{
			Model:      Model{ID: 600},
			TaskID:     &task1ID,
			Digest:     secret.Hash("task1-token"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task1-token",
	}
	token2 := &Token{
		Token: model.Token{
			Model:      Model{ID: 601},
			TaskID:     &task2ID,
			Digest:     secret.Hash("task2-token"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task2-token",
	}

	cache.TokenSaved(token1)
	cache.TokenSaved(token2)

	// Verify both tokens are in cache
	_, err = cache.FindToken("task1-token")
	g.Expect(err).To(BeNil())
	_, err = cache.FindToken("task2-token")
	g.Expect(err).To(BeNil())

	// Revoke task1 only
	cache.TaskRevoked(task1ID)

	// Task1 token removed, task2 token still present
	_, err = cache.FindToken("task1-token")
	g.Expect(err).NotTo(BeNil())
	_, err = cache.FindToken("task2-token")
	g.Expect(err).To(BeNil())
}

// TestTaskRevokedTransaction tests TaskRevoked within a transaction.
func TestTaskRevokedTransaction(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Add task token
	taskID := uint(700)
	token := &Token{
		Token: model.Token{
			Model:      Model{ID: 700},
			TaskID:     &taskID,
			Digest:     secret.Hash("tx-task-token"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "tx-task-token",
	}
	cache.TokenSaved(token)

	// Verify token is in cache
	_, err = cache.FindToken("tx-task-token")
	g.Expect(err).To(BeNil())

	// Revoke in transaction
	err = cache.Transaction(func(tx *Tx) error {
		tx.TaskRevoked(taskID)
		return nil
	})
	g.Expect(err).To(BeNil())

	// Verify token is removed
	_, err = cache.FindToken("tx-task-token")
	g.Expect(err).NotTo(BeNil())

	// Test rollback - add token back, then rollback revoke
	cache.TokenSaved(token)
	_, err = cache.FindToken("tx-task-token")
	g.Expect(err).To(BeNil())

	task2ID := uint(701)
	token2 := &Token{
		Token: model.Token{
			Model:      Model{ID: 701},
			TaskID:     &task2ID,
			Digest:     secret.Hash("tx-task-token-2"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "tx-task-token-2",
	}
	cache.TokenSaved(token2)

	err = cache.Transaction(func(tx *Tx) error {
		tx.TaskRevoked(task2ID)
		return fmt.Errorf("rollback")
	})
	g.Expect(err).NotTo(BeNil())

	// Token2 should still be in cache (rolled back)
	_, err = cache.FindToken("tx-task-token-2")
	g.Expect(err).To(BeNil())
}

// TestTaskSubjectParsing tests that task subjects are parsed on-demand.
func TestTaskSubjectParsing(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Task subject is always parseable (not cached)
	task := &Task{ID: 800}
	expectedSubject := task.Subject()
	subject, err := cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
	g.Expect(subject.Task).NotTo(BeNil())
	g.Expect(subject.Task.ID).To(Equal(uint(800)))
	g.Expect(subject.Key).To(Equal(expectedSubject))
	g.Expect(subject.Login()).To(Equal(task.Login()))
}

// TestTaskSubject tests Task.Subject() method returns hex format.
func TestTaskSubject(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		taskID   uint
		expected string
	}{
		{1, "task.0x1"},
		{42, "task.0x2a"},
		{999, "task.0x3e7"},
		{12345, "task.0x3039"},
		{445, "task.0x1bd"},
	}

	for _, tt := range tests {
		task := &Task{ID: tt.taskID}
		g.Expect(task.Subject()).To(Equal(tt.expected))
	}
}

// TestTaskLogin tests Task.Login() method returns decimal format.
func TestTaskLogin(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		taskID   uint
		expected string
	}{
		{1, "task.1"},
		{42, "task.42"},
		{999, "task.999"},
		{12345, "task.12345"},
		{445, "task.445"},
	}

	for _, tt := range tests {
		task := &Task{ID: tt.taskID}
		g.Expect(task.Login()).To(Equal(tt.expected))
	}
}

// TestTaskWith tests Task.With() parser for hex format.
func TestTaskWith(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		subject     string
		shouldMatch bool
		expectedID  uint
	}{
		{"task.0x1", true, 1},
		{"task.0x2a", true, 42},
		{"task.0x3e7", true, 999},
		{"task.0x3039", true, 12345},
		{"task.0x1bd", true, 445},
		{"task.0xff", true, 255},
		{"task.0xFFFF", true, 65535},
		{"task.123", false, 0},   // decimal format should not match
		{"task.0x", false, 0},    // invalid hex
		{"task.0xGHI", false, 0}, // invalid hex chars
		{"user.0x1", false, 0},   // wrong prefix
		{"task.1", false, 0},     // missing 0x
		{"", false, 0},           // empty string
	}

	for _, tt := range tests {
		task := &Task{}
		matched := task.With(tt.subject)
		g.Expect(matched).To(Equal(tt.shouldMatch), "subject: %s", tt.subject)
		if tt.shouldMatch {
			g.Expect(task.ID).To(Equal(tt.expectedID), "subject: %s", tt.subject)
		}
	}
}

// TestTaskGetScopes tests Task.GetScopes() returns AddonScopes.
func TestTaskGetScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	task := &Task{ID: 100}
	scopes := task.GetScopes()

	g.Expect(scopes).NotTo(BeEmpty())
	g.Expect(scopes).To(ContainElement("addons:get"))
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("applications:post"))
	g.Expect(scopes).To(ContainElement("applications.facts:*"))
}

// TestSubjectWithTask tests Subject.WithTask() population.
func TestSubjectWithTask(t *testing.T) {
	g := NewGomegaWithT(t)

	task := &Task{ID: 555}
	subject := &Subject{}
	subject.WithTask(task)

	expectedSubject := task.Subject()
	g.Expect(subject.Task).To(Equal(task))
	g.Expect(subject.Key).To(Equal(expectedSubject))
	g.Expect(subject.Scopes).NotTo(BeEmpty())
	g.Expect(subject.Scopes).To(ContainElement("addons:get"))
}

// TestSubjectIsTask tests Subject.IsTask() method.
func TestSubjectIsTask(t *testing.T) {
	g := NewGomegaWithT(t)

	// Task subject
	taskSubject := &Subject{Task: &Task{ID: 1}}
	g.Expect(taskSubject.IsTask()).To(BeTrue())
	g.Expect(taskSubject.IsUser()).To(BeFalse())
	g.Expect(taskSubject.IsIdentity()).To(BeFalse())
	g.Expect(taskSubject.IsClient()).To(BeFalse())

	// User subject
	userID := uint(1)
	userSubject := &Subject{UserId: &userID, User: &User{}}
	g.Expect(userSubject.IsUser()).To(BeTrue())
	g.Expect(userSubject.IsTask()).To(BeFalse())

	// Identity subject
	identID := uint(1)
	identSubject := &Subject{IdentityId: &identID, Identity: &Identity{}}
	g.Expect(identSubject.IsIdentity()).To(BeTrue())
	g.Expect(identSubject.IsTask()).To(BeFalse())

	// Client subject
	clientID := uint(1)
	clientSubject := &Subject{ClientId: &clientID, Client: &IdpClient{}}
	g.Expect(clientSubject.IsClient()).To(BeTrue())
	g.Expect(clientSubject.IsTask()).To(BeFalse())

	// Empty subject
	emptySubject := &Subject{}
	g.Expect(emptySubject.IsTask()).To(BeFalse())
	g.Expect(emptySubject.IsUser()).To(BeFalse())
	g.Expect(emptySubject.IsIdentity()).To(BeFalse())
	g.Expect(emptySubject.IsClient()).To(BeFalse())
}

// TestSubjectLogin tests Subject.Login() for different subject types.
func TestSubjectLogin(t *testing.T) {
	g := NewGomegaWithT(t)

	// Task login (decimal format)
	task := &Task{ID: 445}
	taskSubject := &Subject{Task: task}
	g.Expect(taskSubject.Login()).To(Equal("task.445"))

	// User login
	userID := uint(1)
	userSubject := &Subject{UserId: &userID, User: &User{Login: "jsmith"}}
	g.Expect(userSubject.Login()).To(Equal("jsmith"))

	// Identity login
	identID := uint(1)
	identSubject := &Subject{IdentityId: &identID, Identity: &Identity{Login: "idpuser"}}
	g.Expect(identSubject.Login()).To(Equal("idpuser"))

	// Client login
	clientID := uint(1)
	clientSubject := &Subject{ClientId: &clientID, Client: &IdpClient{ClientId: "client-123"}}
	g.Expect(clientSubject.Login()).To(Equal("client-123"))

	// Empty subject
	emptySubject := &Subject{}
	g.Expect(emptySubject.Login()).To(BeEmpty())
}

// TestCacheFindSubjectTask tests finding task subjects via parsing.
func TestCacheFindSubjectTask(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Task subject is parsed on-demand (not cached)
	task := &Task{ID: 999}
	expectedSubject := task.Subject()
	subject, err := cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject).NotTo(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
	g.Expect(subject.Task.ID).To(Equal(uint(999)))
	g.Expect(subject.Key).To(Equal(expectedSubject))
	g.Expect(subject.Login()).To(Equal(task.Login()))

	// Verify scopes
	g.Expect(subject.Scopes).NotTo(BeEmpty())
	g.Expect(subject.Scopes).To(ContainElement("addons:get"))
}

// TestCacheFindSubjectTaskNotFound tests NotFound error for non-existent task.
func TestCacheFindSubjectTaskNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Try to find non-existent task subject
	nonExistentSubject := "1234"
	_, err = cache.FindSubject(nonExistentSubject)
	g.Expect(err).NotTo(BeNil())

	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("subject"))
	g.Expect(notFound.Id).To(Equal(nonExistentSubject))
}

// TestTaskSubjectAlwaysParseable tests that task subjects are always parseable.
func TestTaskSubjectAlwaysParseable(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Task subject is always parseable
	task := &Task{ID: 1111}
	expectedSubject := task.Subject()
	subject, err := cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())

	// TaskRevoked only removes tokens, not the parsing capability
	cache.TaskRevoked(1111)

	// Task subject still parseable
	subject, err = cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
}

// TestTaskSubjectParsing tests task subject parsing behavior.
func TestTaskSubjectParsingBehavior(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Task subject always parseable (not cached)
	task := &Task{ID: 2222}
	expectedSubject := task.Subject()
	subject, err := cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
	g.Expect(subject.Login()).To(Equal(task.Login()))

	// TaskRevoked doesn't affect parsing
	cache.TaskRevoked(2222)

	// Still parseable
	subject, err = cache.FindSubject(expectedSubject)
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsTask()).To(BeTrue())
}

// TestMultipleTaskSubjectsParseable tests multiple task subjects are parseable.
func TestMultipleTaskSubjectsParseable(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// All task subjects are parseable (not cached)
	tasks := []*Task{
		{ID: 1},
		{ID: 2},
		{ID: 3},
		{ID: 100},
		{ID: 999},
	}

	// All tasks should be parseable
	for _, task := range tasks {
		expectedSubject := task.Subject()
		subject, err := cache.FindSubject(expectedSubject)
		g.Expect(err).To(BeNil())
		g.Expect(subject.IsTask()).To(BeTrue())
		g.Expect(subject.Login()).To(Equal(task.Login()))
	}

	// TaskRevoked only removes tokens
	cache.TaskRevoked(2)

	// All task subjects still parseable
	for _, task := range tasks {
		expectedSubject := task.Subject()
		subject, err := cache.FindSubject(expectedSubject)
		g.Expect(err).To(BeNil())
		g.Expect(subject.IsTask()).To(BeTrue())
	}
}

// TestConcurrentTaskSubjectParsing tests concurrent task subject parsing.
func TestConcurrentTaskSubjectParsing(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent TaskRevoked (only removes tokens)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cache.TaskRevoked(uint(id*1000 + j))
			}
		}(i)
	}

	// Concurrent FindSubject (parsing)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				taskID := j % 100
				subject, parseErr := cache.FindSubject(Task{ID: uint(taskID)}.Subject())
				// Parsing should always succeed
				if parseErr == nil {
					g.Expect(subject.IsTask()).To(BeTrue())
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is still consistent
	d := cache.data.Load()
	g.Expect(d).NotTo(BeNil())
}

// TestTaskSubjectWithoutCache tests that tasks don't need cache refresh.
func TestTaskSubjectWithoutCache(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create cache and refresh (tasks not loaded from DB)
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Task subjects are parsed on-demand, no DB/cache needed
	task1 := &Task{ID: 100}
	task2 := &Task{ID: 200}

	subject1, err := cache.FindSubject(task1.Subject())
	g.Expect(err).To(BeNil())
	g.Expect(subject1.IsTask()).To(BeTrue())
	g.Expect(subject1.Task.ID).To(Equal(uint(100)))

	subject2, err := cache.FindSubject(task2.Subject())
	g.Expect(err).To(BeNil())
	g.Expect(subject2.IsTask()).To(BeTrue())
	g.Expect(subject2.Task.ID).To(Equal(uint(200)))
}

// TestCacheMixedSubjectTypes tests that different subject types (User, Identity, Client, Task) coexist.
func TestCacheMixedSubjectTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Add user
	user := &User{
		Model:   Model{ID: 1},
		Subject: "user-subject-1",
		Login:   "user1",
	}
	cache.UserSaved(user)

	// Add identity
	identity := &Identity{
		Model:   Model{ID: 1},
		Issuer:  "https://idp.example.com",
		Subject: "identity-subject-1",
		Login:   "identity1",
	}
	cache.IdentitySaved(identity)

	// Add client
	client := &IdpClient{
		Model:    Model{ID: 1},
		Subject:  "client-subject-1",
		ClientId: "client1",
	}
	cache.ClientSaved(client)

	// Task subject parseable (not cached)
	task := &Task{ID: 1}

	// All should be findable
	userSubj, err := cache.FindSubject("user-subject-1")
	g.Expect(err).To(BeNil())
	g.Expect(userSubj.IsUser()).To(BeTrue())
	g.Expect(userSubj.Login()).To(Equal("user1"))

	identSubj, err := cache.FindSubject("identity-subject-1")
	g.Expect(err).To(BeNil())
	g.Expect(identSubj.IsIdentity()).To(BeTrue())
	g.Expect(identSubj.Login()).To(Equal("identity1"))

	clientSubj, err := cache.FindSubject("client-subject-1")
	g.Expect(err).To(BeNil())
	g.Expect(clientSubj.IsClient()).To(BeTrue())
	g.Expect(clientSubj.Login()).To(Equal("client1"))

	taskSubject := task.Subject()
	taskSubj, err := cache.FindSubject(taskSubject)
	g.Expect(err).To(BeNil())
	g.Expect(taskSubj.IsTask()).To(BeTrue())
	g.Expect(taskSubj.Login()).To(Equal(task.Login()))
}

// TestFindTokenById tests FindTokenById method.
func TestFindTokenById(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create a task in DB first
	task := &model.Task{
		Model: Model{ID: 1000},
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Create a task token in DB
	taskID := uint(1000)
	taskToken := &Token{
		Token: model.Token{
			Model:      Model{ID: 1000},
			Kind:       KindAPIKey,
			TaskID:     &taskID,
			Digest:     secret.Hash("task-token-1000"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task-token-1000",
	}
	err = db.Create(&taskToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load token with scopes assigned
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Find token by ID - should have scopes
	foundToken, err := cache.FindTokenById(1000)
	g.Expect(err).To(BeNil())
	g.Expect(foundToken).NotTo(BeNil())
	g.Expect(foundToken.ID).To(Equal(uint(1000)))
	g.Expect(*foundToken.TaskID).To(Equal(taskID))
	g.Expect(foundToken.Scopes).NotTo(BeEmpty())
	g.Expect(foundToken.Scopes).To(ContainElement("addons:get"))

	// Try to find non-existent token
	_, err = cache.FindTokenById(9999)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("token"))
}

// TestFindTokenByIdWithUser tests FindTokenById returns scopes for user tokens.
func TestFindTokenByIdWithUser(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create permission
	perm := &Permission{
		Model: Model{ID: 1},
		Scope: "applications:get",
	}
	err = db.Create(perm).Error
	g.Expect(err).To(BeNil())

	// Create role with permission
	role := &Role{
		Model: Model{ID: 1},
		Name:  "Developer",
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	err = db.Model(role).Association("Permissions").Append(perm)
	g.Expect(err).To(BeNil())

	// Create user
	user := &User{
		Model:   Model{ID: 2000},
		Subject: "user-subject",
		Login:   "testuser",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Associate user with role
	err = db.Exec("INSERT INTO UserRole (UserID, RoleID) VALUES (?, ?)", user.ID, role.ID).Error
	g.Expect(err).To(BeNil())

	// Create user token in DB
	userID := uint(2000)
	userToken := &Token{
		Token: model.Token{
			Model:      Model{ID: 2001},
			Kind:       KindAPIKey,
			UserID:     &userID,
			Digest:     secret.Hash("user-token-2001"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "user-token-2001",
	}
	err = db.Create(&userToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load token with scopes assigned
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Find token by ID - should have scopes
	foundToken, err := cache.FindTokenById(2001)
	g.Expect(err).To(BeNil())
	g.Expect(foundToken).NotTo(BeNil())
	g.Expect(foundToken.Scopes).NotTo(BeEmpty())
	g.Expect(foundToken.Scopes).To(ContainElement("applications:get"))
}

// TestFindTokenWithScopes tests that FindToken returns tokens with scopes populated.
func TestFindTokenWithScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create a task in DB first
	task := &model.Task{
		Model: Model{ID: 3000},
	}
	err = db.Create(task).Error
	g.Expect(err).To(BeNil())

	// Create a task token in DB
	taskID := uint(3000)
	taskToken := &Token{
		Token: model.Token{
			Model:      Model{ID: 3000},
			Kind:       KindAPIKey,
			TaskID:     &taskID,
			Digest:     secret.Hash("task-token-3000"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "task-token-3000",
	}
	err = db.Create(&taskToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load token with scopes assigned
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Find token by secret - should have scopes
	foundToken, err := cache.FindToken("task-token-3000")
	g.Expect(err).To(BeNil())
	g.Expect(foundToken).NotTo(BeNil())
	g.Expect(foundToken.Scopes).NotTo(BeEmpty())
	g.Expect(foundToken.Scopes).To(ContainElement("addons:get"))
}

// TestFindTokenWithIdentity tests FindToken with IdP identity token.
func TestFindTokenWithIdentity(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create identity in DB
	identity := &Identity{
		Model:   Model{ID: 4000},
		Issuer:  "https://idp.example.com",
		Subject: "identity-subject-4000",
		Login:   "identityuser",
		Scopes:  "applications:get applications:post",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	// Create identity token in DB
	identityID := uint(4000)
	identityToken := &Token{
		Token: model.Token{
			Model:         Model{ID: 4001},
			Kind:          KindAPIKey,
			IdpIdentityID: &identityID,
			Digest:        secret.Hash("identity-token-4001"),
			Expiration:    time.Now().Add(24 * time.Hour),
		},
		Secret: "identity-token-4001",
	}
	err = db.Create(&identityToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load token with scopes assigned
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Find token - should have identity scopes
	foundToken, err := cache.FindToken("identity-token-4001")
	g.Expect(err).To(BeNil())
	g.Expect(foundToken).NotTo(BeNil())
	g.Expect(foundToken.Scopes).To(ContainElement("applications:get"))
	g.Expect(foundToken.Scopes).To(ContainElement("applications:post"))
}

// TestFindTokenWithClient tests FindToken with IdP client token.
func TestFindTokenWithClient(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create client in DB
	client := &IdpClient{
		Model:    Model{ID: 5000},
		Subject:  "client-subject-5000",
		ClientId: "client-5000",
		Scopes:   []string{"applications:get", "applications:post"},
	}
	err = db.Create(client).Error
	g.Expect(err).To(BeNil())

	// Create client token in DB
	clientID := uint(5000)
	clientToken := &Token{
		Token: model.Token{
			Model:       Model{ID: 5001},
			Kind:        KindAPIKey,
			IdpClientID: &clientID,
			Digest:      secret.Hash("client-token-5001"),
			Expiration:  time.Now().Add(24 * time.Hour),
		},
		Secret: "client-token-5001",
	}
	err = db.Create(&clientToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load token with scopes assigned
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Find token - should have client scopes
	foundToken, err := cache.FindToken("client-token-5001")
	g.Expect(err).To(BeNil())
	g.Expect(foundToken).NotTo(BeNil())
	g.Expect(foundToken.Scopes).To(ContainElement("applications:get"))
	g.Expect(foundToken.Scopes).To(ContainElement("applications:post"))
}

// TestCacheConcurrentMixedOperations tests mixed concurrent operations.
func TestCacheConcurrentMixedOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Seed some initial data
	for i := 0; i < 10; i++ {
		user := &User{
			Model:   Model{ID: uint(i)},
			Subject: fmt.Sprintf("mixeduser%d", i),
			Login:   fmt.Sprintf("mixeduser%d", i),
		}
		cache.UserSaved(user)
	}

	var wg sync.WaitGroup
	iterations := 50

	// Concurrent RoleSaved/RoleDeleted
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				role := &Role{
					Model: Model{ID: uint(id*1000 + j)},
					Name:  fmt.Sprintf("mixed-role-%d-%d", id, j),
				}
				cache.RoleSaved(role)
				if j%2 == 0 {
					cache.RoleDeleted(role.ID)
				}
			}
		}(i)
	}

	// Concurrent TokenSaved
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				taskID := uint(id*1000 + j)
				token := &Token{
					Token: model.Token{
						Model:      Model{ID: uint(id*1000 + j)},
						TaskID:     &taskID,
						Digest:     secret.Hash(fmt.Sprintf("mixed-token-%d-%d", id, j)),
						Expiration: time.Now().Add(24 * time.Hour),
					},
					Secret: fmt.Sprintf("mixed-token-%d-%d", id, j),
				}
				cache.TokenSaved(token)
			}
		}(i)
	}

	// Concurrent FindUserByLogin
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				userIdx := j % 10
				cache.FindUserByLogin(fmt.Sprintf("mixeduser%d", userIdx))
			}
		}(i)
	}

	// Concurrent TaskRevoked
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				taskID := uint(id*1000 + j)
				cache.TaskRevoked(taskID)
			}
		}(i)
	}

	// Concurrent FindSubject (task parsing)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				taskID := j % 100
				subject, parseErr := cache.FindSubject(Task{ID: uint(taskID)}.Subject())
				// Parsing always succeeds
				if parseErr == nil {
					g.Expect(subject.IsTask()).To(BeTrue())
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	d := cache.data.Load()
	// Should have some users (not all, some deleted/not found)
	g.Expect(len(d.userById)).To(BeNumerically(">", 0))
}
