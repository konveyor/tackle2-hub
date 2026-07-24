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
		&ServiceAccount{},
		&model.Task{},
		&Role{},
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

// TestSaSavedDeleted tests SaSaved/SaDeleted methods.
func TestSaSavedDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	sa := &ServiceAccount{
		Model:   Model{ID: 100},
		Subject: "sa-subject-100",
		Name:    "ci-bot",
	}
	cache.SaSaved(sa)

	// Findable by subject.
	subject, err := cache.FindSubject("sa-subject-100")
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsServiceAccount()).To(BeTrue())
	g.Expect(subject.IsUser()).To(BeFalse())
	g.Expect(subject.ServiceAccountId).NotTo(BeNil())
	g.Expect(*subject.ServiceAccountId).To(Equal(uint(100)))
	g.Expect(subject.ServiceAccount).NotTo(BeNil())
	g.Expect(subject.ServiceAccount.Name).To(Equal("ci-bot"))
	g.Expect(subject.Key).To(Equal("sa-subject-100"))

	// Delete removes from cache.
	cache.SaDeleted(100)
	_, err = cache.FindSubject("sa-subject-100")
	g.Expect(err).NotTo(BeNil())
}

// TestSaDeletedRemovesTokens tests that SaDeleted removes associated tokens.
func TestSaDeletedRemovesTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	sa := &ServiceAccount{
		Model:   Model{ID: 200},
		Subject: "sa-subject-200",
		Name:    "deploy-bot",
	}
	cache.SaSaved(sa)

	// Add tokens for this service account.
	saID := uint(200)
	token1 := &Token{
		Token: model.Token{
			Model:            Model{ID: 201},
			ServiceAccountID: &saID,
			Digest:           secret.Hash("sa-token-201"),
			Expiration:       time.Now().Add(24 * time.Hour),
		},
		Secret: "sa-token-201",
	}
	token2 := &Token{
		Token: model.Token{
			Model:            Model{ID: 202},
			ServiceAccountID: &saID,
			Digest:           secret.Hash("sa-token-202"),
			Expiration:       time.Now().Add(24 * time.Hour),
		},
		Secret: "sa-token-202",
	}
	cache.TokenSaved(token1)
	cache.TokenSaved(token2)

	_, err = cache.FindToken("sa-token-201")
	g.Expect(err).To(BeNil())
	_, err = cache.FindToken("sa-token-202")
	g.Expect(err).To(BeNil())

	// Delete SA should remove both tokens.
	cache.SaDeleted(200)

	_, err = cache.FindToken("sa-token-201")
	g.Expect(err).NotTo(BeNil())
	_, err = cache.FindToken("sa-token-202")
	g.Expect(err).NotTo(BeNil())
}

// TestSaDeletedDoesNotRemoveUserTokens tests that SaDeleted does not remove user tokens.
func TestSaDeletedDoesNotRemoveUserTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// SA and user with same numeric ID.
	saID := uint(300)
	sa := &ServiceAccount{
		Model:   Model{ID: saID},
		Subject: "sa-subject-300",
		Name:    "bot",
	}
	cache.SaSaved(sa)

	user := &User{
		Model:   Model{ID: saID},
		Subject: "user-subject-300",
		Login:   "user300",
	}
	cache.UserSaved(user)

	// Token belonging to user.
	userToken := &Token{
		Token: model.Token{
			Model:      Model{ID: 301},
			UserID:     &saID,
			Digest:     secret.Hash("user-token-301"),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: "user-token-301",
	}
	cache.TokenSaved(userToken)

	// Token belonging to SA.
	saToken := &Token{
		Token: model.Token{
			Model:            Model{ID: 302},
			ServiceAccountID: &saID,
			Digest:           secret.Hash("sa-token-302"),
			Expiration:       time.Now().Add(24 * time.Hour),
		},
		Secret: "sa-token-302",
	}
	cache.TokenSaved(saToken)

	// Delete SA.
	cache.SaDeleted(saID)

	// User token should still exist.
	_, err = cache.FindToken("user-token-301")
	g.Expect(err).To(BeNil())

	// SA token should be removed.
	_, err = cache.FindToken("sa-token-302")
	g.Expect(err).NotTo(BeNil())
}

// TestSaScopes tests that service account scopes are derived from roles.
func TestSaScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create roles in DB.
	role1 := &Role{
		Model:  Model{ID: 1},
		Name:   "admin",
		Scopes: []string{"applications:get", "applications:post"},
	}
	role2 := &Role{
		Model:  Model{ID: 2},
		Name:   "viewer",
		Scopes: []string{"applications:get"},
	}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(role2).Error
	g.Expect(err).To(BeNil())

	// Create SA with roles in DB.
	sa := &ServiceAccount{
		Model:   Model{ID: 400},
		Subject: "sa-subject-400",
		Name:    "scoped-bot",
		Roles:   []Role{*role1, *role2},
	}
	err = db.Create(sa).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	subject, err := cache.FindSubject("sa-subject-400")
	g.Expect(err).To(BeNil())
	g.Expect(subject.IsServiceAccount()).To(BeTrue())
	g.Expect(subject.Scopes).To(ContainElement("applications:get"))
	g.Expect(subject.Scopes).To(ContainElement("applications:post"))
}

// TestSaLogin tests Subject.Login() for service account subjects.
func TestSaLogin(t *testing.T) {
	g := NewGomegaWithT(t)

	saID := uint(1)
	subject := &Subject{
		ServiceAccountId: &saID,
		ServiceAccount: &ServiceAccount{
			Name: "ci-bot",
		},
	}
	g.Expect(subject.IsServiceAccount()).To(BeTrue())
	g.Expect(subject.Login()).To(Equal("ci-bot"))
}

// TestSubjectIsServiceAccount tests Subject.IsServiceAccount().
func TestSubjectIsServiceAccount(t *testing.T) {
	g := NewGomegaWithT(t)

	// SA subject.
	saID := uint(1)
	saSubject := &Subject{ServiceAccountId: &saID, ServiceAccount: &ServiceAccount{}}
	g.Expect(saSubject.IsServiceAccount()).To(BeTrue())
	g.Expect(saSubject.IsUser()).To(BeFalse())

	// Empty subject.
	emptySubject := &Subject{}
	g.Expect(emptySubject.IsServiceAccount()).To(BeFalse())
}

// TestSaMixedSubjectTypes tests that SA subjects coexist with other subject types.
func TestSaMixedSubjectTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	user := &User{
		Model:   Model{ID: 1},
		Subject: "user-subject-1",
		Login:   "user1",
	}
	cache.UserSaved(user)

	sa := &ServiceAccount{
		Model:   Model{ID: 2},
		Subject: "sa-subject-2",
		Name:    "ci-bot",
	}
	cache.SaSaved(sa)

	// Both findable.
	userSubj, err := cache.FindSubject("user-subject-1")
	g.Expect(err).To(BeNil())
	g.Expect(userSubj.IsUser()).To(BeTrue())
	g.Expect(userSubj.IsServiceAccount()).To(BeFalse())

	saSubj, err := cache.FindSubject("sa-subject-2")
	g.Expect(err).To(BeNil())
	g.Expect(saSubj.IsServiceAccount()).To(BeTrue())
	g.Expect(saSubj.IsUser()).To(BeFalse())
	g.Expect(saSubj.Login()).To(Equal("ci-bot"))
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

// TestSubjectLogin tests Subject.Login() for different subject types.
func TestSubjectLogin(t *testing.T) {
	g := NewGomegaWithT(t)

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

	// Create role with scopes
	role := &Role{
		Model:  Model{ID: 1},
		Name:   "Developer",
		Scopes: []string{"applications:get"},
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	// Create user
	user := &User{
		Model:   Model{ID: 2000},
		Subject: "user-subject",
		Login:   "testuser",
		Roles:   []Role{*role},
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// Create user token in DB
	userID := uint(2000)
	userToken := &Token{
		Token: model.Token{
			Model:      Model{ID: 2001},
			Kind:       KindAPIKey,
			Subject:    user.Subject,
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
			Scopes:        []string{"applications:get", "applications:post"},
		},
		Secret: "identity-token-4001",
	}
	err = db.Create(&identityToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache to load token with scopes
	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Find token - should have scopes
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

	wg.Wait()

	// Verify final state is consistent
	d := cache.data.Load()
	// Should have some users (not all, some deleted/not found)
	g.Expect(len(d.userById)).To(BeNumerically(">", 0))
}

// TestGrantDeletedCascadesToTokens tests that deleting a grant cascades to associated tokens.
func TestGrantDeletedCascadesToTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create IdpIdentity
	identity := &Identity{
		Model:   Model{ID: 4000},
		Issuer:  "https://idp.example.com",
		Subject: "identity-subject-4000",
		Login:   "cascadeuser",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())
	cache.IdentitySaved(identity)

	identityID := uint(4000)

	// Create grant
	grant := &Grant{
		Model:         Model{ID: 4001, UpdateTime: time.Now()},
		Kind:          KindAuthCode,
		ClientId:      "test-client",
		AuthId:        "auth-4001",
		Subject:       "identity-subject-4000",
		RefreshToken:  secret.Hash("refresh-4001"),
		IdpIdentityID: &identityID,
		Expiration:    time.Now().Add(24 * time.Hour),
	}
	err = db.Create(grant).Error
	g.Expect(err).To(BeNil())

	// Create multiple tokens for this grant
	grantID := uint(4001)
	token1 := &Token{
		Token: model.Token{
			Model:         Model{ID: 4002},
			Kind:          KindAPIKey,
			AuthId:        "auth-token-4002",
			Subject:       "identity-subject-4000",
			IdpIdentityID: &identityID,
			GrantID:       &grantID,
			Digest:        secret.Hash("cascade-token-1"),
			Expiration:    time.Now().Add(24 * time.Hour),
		},
		Secret: "cascade-token-1",
	}
	token2 := &Token{
		Token: model.Token{
			Model:         Model{ID: 4003},
			Kind:          KindAPIKey,
			AuthId:        "auth-token-4003",
			Subject:       "identity-subject-4000",
			IdpIdentityID: &identityID,
			GrantID:       &grantID,
			Digest:        secret.Hash("cascade-token-2"),
			Expiration:    time.Now().Add(24 * time.Hour),
		},
		Secret: "cascade-token-2",
	}

	err = db.Create(&token1.Token).Error
	g.Expect(err).To(BeNil())
	cache.TokenSaved(token1)

	err = db.Create(&token2.Token).Error
	g.Expect(err).To(BeNil())
	cache.TokenSaved(token2)

	// Verify both tokens exist
	_, err = cache.FindToken("cascade-token-1")
	g.Expect(err).To(BeNil())
	_, err = cache.FindToken("cascade-token-2")
	g.Expect(err).To(BeNil())

	// Delete grant - should cascade delete both tokens
	cache.GrantDeleted(4001)

	// Both tokens should be deleted
	_, err = cache.FindToken("cascade-token-1")
	g.Expect(err).NotTo(BeNil())
	_, err = cache.FindToken("cascade-token-2")
	g.Expect(err).NotTo(BeNil())
}

// TestFindClientById tests finding clients by ID.
func TestFindClientById(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create test client
	client := &IdpClient{
		Model:           Model{ID: 100},
		Subject:         "test-client-subject",
		ClientId:        "test-client-id",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid", "profile"},
	}
	cache.ClientSaved(client)

	// Test finding client by ID - should succeed
	found, err := cache.FindClientById(100)
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.ID).To(Equal(uint(100)))
	g.Expect(found.Subject).To(Equal("test-client-subject"))
	g.Expect(found.ClientId).To(Equal("test-client-id"))
	g.Expect(found.ApplicationType).To(Equal("web"))
	g.Expect(found.Grants).To(Equal([]string{"client_credentials"}))
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile"}))

	// Test finding non-existent client by ID - should fail
	_, err = cache.FindClientById(999)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("client"))
	g.Expect(notFound.Id).To(Equal("999"))
}

// TestFindClientBySubject tests finding clients by subject.
func TestFindClientBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create test client
	client := &IdpClient{
		Model:           Model{ID: 200},
		Subject:         "client-subject-200",
		ClientId:        "client-id-200",
		ApplicationType: "native",
		Grants:          []string{"authorization_code"},
		Scopes:          []string{"openid", "email"},
	}
	cache.ClientSaved(client)

	// Test finding client by subject - should succeed
	found, err := cache.FindClientBySubject("client-subject-200")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.ID).To(Equal(uint(200)))
	g.Expect(found.Subject).To(Equal("client-subject-200"))
	g.Expect(found.ClientId).To(Equal("client-id-200"))
	g.Expect(found.ApplicationType).To(Equal("native"))
	g.Expect(found.Grants).To(Equal([]string{"authorization_code"}))
	g.Expect(found.Scopes).To(Equal([]string{"openid", "email"}))

	// Test finding non-existent client by subject - should fail
	_, err = cache.FindClientBySubject("non-existent-subject")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("client"))
	g.Expect(notFound.Id).To(Equal("non-existent-subject"))
}

// TestFindClientMultipleClients tests finding clients when multiple exist.
func TestFindClientMultipleClients(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create multiple test clients
	client1 := &IdpClient{
		Model:           Model{ID: 301},
		Subject:         "client-subject-301",
		ClientId:        "client-id-301",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	client2 := &IdpClient{
		Model:           Model{ID: 302},
		Subject:         "client-subject-302",
		ClientId:        "client-id-302",
		ApplicationType: "native",
		Grants:          []string{"authorization_code"},
		Scopes:          []string{"openid", "profile"},
	}
	client3 := &IdpClient{
		Model:           Model{ID: 303},
		Subject:         "client-subject-303",
		ClientId:        "client-id-303",
		ApplicationType: "web",
		Grants:          []string{"client_credentials", "refresh_token"},
		Scopes:          []string{"openid", "profile", "email"},
	}

	cache.ClientSaved(client1)
	cache.ClientSaved(client2)
	cache.ClientSaved(client3)

	// Find each client by ID
	found1, err := cache.FindClientById(301)
	g.Expect(err).To(BeNil())
	g.Expect(found1.Subject).To(Equal("client-subject-301"))

	found2, err := cache.FindClientById(302)
	g.Expect(err).To(BeNil())
	g.Expect(found2.Subject).To(Equal("client-subject-302"))

	found3, err := cache.FindClientById(303)
	g.Expect(err).To(BeNil())
	g.Expect(found3.Subject).To(Equal("client-subject-303"))

	// Find each client by subject
	found1, err = cache.FindClientBySubject("client-subject-301")
	g.Expect(err).To(BeNil())
	g.Expect(found1.ID).To(Equal(uint(301)))

	found2, err = cache.FindClientBySubject("client-subject-302")
	g.Expect(err).To(BeNil())
	g.Expect(found2.ID).To(Equal(uint(302)))

	found3, err = cache.FindClientBySubject("client-subject-303")
	g.Expect(err).To(BeNil())
	g.Expect(found3.ID).To(Equal(uint(303)))
}

// TestClientDeleted tests that deleting a client removes it from cache.
func TestClientDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create and save client
	client := &IdpClient{
		Model:           Model{ID: 400},
		Subject:         "client-subject-400",
		ClientId:        "client-id-400",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	cache.ClientSaved(client)

	// Verify client exists
	found, err := cache.FindClientById(400)
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(400)))

	found, err = cache.FindClientBySubject("client-subject-400")
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(400)))

	// Delete client
	cache.ClientDeleted(400)

	// Verify client is removed from both indexes
	_, err = cache.FindClientById(400)
	g.Expect(err).NotTo(BeNil())

	_, err = cache.FindClientBySubject("client-subject-400")
	g.Expect(err).NotTo(BeNil())
}

// TestClientSavedUpdate tests that saving an existing client updates the cache.
func TestClientSavedUpdate(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create and save initial client
	client := &IdpClient{
		Model:           Model{ID: 500},
		Subject:         "client-subject-500",
		ClientId:        "client-id-500",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	cache.ClientSaved(client)

	// Verify initial state
	found, err := cache.FindClientById(500)
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid"}))

	// Update and save client with new scopes
	client.Scopes = []string{"openid", "profile", "email"}
	cache.ClientSaved(client)

	// Verify updated state
	found, err = cache.FindClientById(500)
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile", "email"}))

	// Verify can still find by subject
	found, err = cache.FindClientBySubject("client-subject-500")
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile", "email"}))
}

// TestFindClientByStrId tests finding clients by string ClientId.
func TestFindClientByStrId(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create test client
	client := &IdpClient{
		Model:           Model{ID: 600},
		Subject:         "client-subject-600",
		ClientId:        "test-client-string-id",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid", "profile"},
	}
	cache.ClientSaved(client)

	// Test finding client by string ClientId - should succeed
	found, err := cache.FindClientByStrId("test-client-string-id")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.ID).To(Equal(uint(600)))
	g.Expect(found.Subject).To(Equal("client-subject-600"))
	g.Expect(found.ClientId).To(Equal("test-client-string-id"))
	g.Expect(found.ApplicationType).To(Equal("web"))
	g.Expect(found.Grants).To(Equal([]string{"client_credentials"}))
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile"}))

	// Test finding non-existent client by string ClientId - should fail
	_, err = cache.FindClientByStrId("non-existent-client-id")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("client"))
	g.Expect(notFound.Id).To(Equal("non-existent-client-id"))
}

// TestFindClientByStrIdMultipleClients tests finding clients by ClientId with multiple clients.
func TestFindClientByStrIdMultipleClients(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create multiple test clients with different ClientIds
	client1 := &IdpClient{
		Model:           Model{ID: 701},
		Subject:         "client-subject-701",
		ClientId:        "client-credentials-app",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	client2 := &IdpClient{
		Model:           Model{ID: 702},
		Subject:         "client-subject-702",
		ClientId:        "auth-code-app",
		ApplicationType: "native",
		Grants:          []string{"authorization_code"},
		Scopes:          []string{"openid", "profile"},
	}
	client3 := &IdpClient{
		Model:           Model{ID: 703},
		Subject:         "client-subject-703",
		ClientId:        "web-app-client",
		ApplicationType: "web",
		Grants:          []string{"authorization_code", "refresh_token"},
		Scopes:          []string{"openid", "profile", "email"},
	}

	cache.ClientSaved(client1)
	cache.ClientSaved(client2)
	cache.ClientSaved(client3)

	// Find each client by ClientId
	found1, err := cache.FindClientByStrId("client-credentials-app")
	g.Expect(err).To(BeNil())
	g.Expect(found1.ID).To(Equal(uint(701)))
	g.Expect(found1.Subject).To(Equal("client-subject-701"))

	found2, err := cache.FindClientByStrId("auth-code-app")
	g.Expect(err).To(BeNil())
	g.Expect(found2.ID).To(Equal(uint(702)))
	g.Expect(found2.Subject).To(Equal("client-subject-702"))

	found3, err := cache.FindClientByStrId("web-app-client")
	g.Expect(err).To(BeNil())
	g.Expect(found3.ID).To(Equal(uint(703)))
	g.Expect(found3.Subject).To(Equal("client-subject-703"))

	// Verify can also find by numeric ID and subject
	found1ById, err := cache.FindClientById(701)
	g.Expect(err).To(BeNil())
	g.Expect(found1ById.ClientId).To(Equal("client-credentials-app"))

	found1BySubject, err := cache.FindClientBySubject("client-subject-701")
	g.Expect(err).To(BeNil())
	g.Expect(found1BySubject.ClientId).To(Equal("client-credentials-app"))
}

// TestClientDeletedRemovesFromAllIndexes tests that deleting a client removes it from all three indexes.
func TestClientDeletedRemovesFromAllIndexes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create and save client
	client := &IdpClient{
		Model:           Model{ID: 800},
		Subject:         "client-subject-800",
		ClientId:        "deletable-client-id",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	cache.ClientSaved(client)

	// Verify client exists in all three indexes
	found, err := cache.FindClientById(800)
	g.Expect(err).To(BeNil())
	g.Expect(found.ClientId).To(Equal("deletable-client-id"))

	found, err = cache.FindClientByStrId("deletable-client-id")
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(800)))

	found, err = cache.FindClientBySubject("client-subject-800")
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(800)))

	// Delete client
	cache.ClientDeleted(800)

	// Verify client is removed from all three indexes
	_, err = cache.FindClientById(800)
	g.Expect(err).NotTo(BeNil())

	_, err = cache.FindClientByStrId("deletable-client-id")
	g.Expect(err).NotTo(BeNil())

	_, err = cache.FindClientBySubject("client-subject-800")
	g.Expect(err).NotTo(BeNil())
}

// TestClientSavedUpdateAllIndexes tests that updating a client updates all indexes.
func TestClientSavedUpdateAllIndexes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create and save initial client
	client := &IdpClient{
		Model:           Model{ID: 900},
		Subject:         "client-subject-900",
		ClientId:        "updatable-client",
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid"},
	}
	cache.ClientSaved(client)

	// Verify initial state via all indexes
	found, err := cache.FindClientById(900)
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid"}))

	found, err = cache.FindClientByStrId("updatable-client")
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid"}))

	found, err = cache.FindClientBySubject("client-subject-900")
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid"}))

	// Update client with new scopes
	client.Scopes = []string{"openid", "profile", "email"}
	cache.ClientSaved(client)

	// Verify updated state via all indexes
	found, err = cache.FindClientById(900)
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile", "email"}))

	found, err = cache.FindClientByStrId("updatable-client")
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile", "email"}))

	found, err = cache.FindClientBySubject("client-subject-900")
	g.Expect(err).To(BeNil())
	g.Expect(found.Scopes).To(Equal([]string{"openid", "profile", "email"}))
}

// TestClientByStrIdWithSpecialCharacters tests ClientId strings with special characters.
func TestClientByStrIdWithSpecialCharacters(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create clients with special characters in ClientId
	specialClients := []struct {
		id       uint
		clientId string
	}{
		{1001, "client-with-dashes"},
		{1002, "client_with_underscores"},
		{1003, "client.with.dots"},
		{1004, "client@with@at"},
		{1005, "MixedCaseClient"},
	}

	for _, sc := range specialClients {
		client := &IdpClient{
			Model:           Model{ID: sc.id},
			Subject:         fmt.Sprintf("subject-%d", sc.id),
			ClientId:        sc.clientId,
			ApplicationType: "web",
			Grants:          []string{"client_credentials"},
			Scopes:          []string{"openid"},
		}
		cache.ClientSaved(client)
	}

	// Verify all can be found by their ClientId
	for _, sc := range specialClients {
		found, err := cache.FindClientByStrId(sc.clientId)
		g.Expect(err).To(BeNil())
		g.Expect(found.ID).To(Equal(sc.id))
		g.Expect(found.ClientId).To(Equal(sc.clientId))
	}

	// Verify case sensitivity - "MixedCaseClient" != "mixedcaseclient"
	_, err = cache.FindClientByStrId("mixedcaseclient")
	g.Expect(err).NotTo(BeNil())

	found, err := cache.FindClientByStrId("MixedCaseClient")
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(1005)))
}
