package cache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
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
		&Task{},
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

// TestCacheDoubleCheckRefresh tests that concurrent ensureFresh calls don't cause issues.
func TestCacheDoubleCheckRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test data
	user := &User{
		Subject:  "double-check-user",
		Login:    "doublecheckuser",
		Password: secret.HashPassword("password"),
		Email:    "doublecheck@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Force cache to be stale - not needed with new design
	// (background refresh runs independently)

	// Launch multiple concurrent calls to FindUserByLogin (which calls ensureFresh)
	var wg sync.WaitGroup
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := cache.FindUserByLogin("doublecheckuser")
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// All calls should succeed without error
	for i, err := range errors {
		g.Expect(err).To(BeNil(), fmt.Sprintf("goroutine %d failed", i))
	}

	// Cache should have refreshed and found the user
	found, err := cache.FindUserByLogin("doublecheckuser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
}

// TestCacheInconsistency tests error paths when referenced entities are missing.
func TestCacheInconsistency(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create token referencing non-existent user
	userID := uint(9999)
	userTokenSecret := "inconsistent-user-token"
	userTokenDigest := secret.Hash(userTokenSecret)
	userToken := &Token{
		Token: model.Token{
			Model:  Model{ID: 1},
			UserID: &userID,
			Digest: userTokenDigest,
		},
	}
	d := cache.data.Load()
	d.tokenByDigest[userTokenDigest] = userToken
	d.tokenById[1] = userToken

	_, err = cache.FindToken(userTokenSecret)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("user"))

	// Create token referencing non-existent identity
	identityID := uint(7777)
	identityTokenSecret := "inconsistent-identity-token"
	identityTokenDigest := secret.Hash(identityTokenSecret)
	identityToken := &Token{
		Token: model.Token{
			Model:         Model{ID: 3},
			IdpIdentityID: &identityID,
			Digest:        identityTokenDigest,
		},
	}
	d.tokenByDigest[identityTokenDigest] = identityToken
	d.tokenById[3] = identityToken

	_, err = cache.FindToken(identityTokenSecret)
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("identity"))

	// Create token referencing non-existent client
	clientID := uint(6666)
	clientTokenSecret := "inconsistent-client-token"
	clientTokenDigest := secret.Hash(clientTokenSecret)
	clientToken := &Token{
		Token: model.Token{
			Model:       Model{ID: 4},
			IdpClientID: &clientID,
			Digest:      clientTokenDigest,
		},
	}
	d.tokenByDigest[clientTokenDigest] = clientToken
	d.tokenById[4] = clientToken

	_, err = cache.FindToken(clientTokenSecret)
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("client"))
}

// TestTaskRevokedRemovesTokens tests that TaskRevoked removes tokens from token cache.
func TestTaskRevokedRemovesTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create task token
	taskID := uint(500)
	tokenSecret := "task-token-secret"
	token := &Token{
		Token: model.Token{
			Model:      Model{ID: 100},
			TaskID:     &taskID,
			Digest:     secret.Hash(tokenSecret),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: tokenSecret,
	}

	// Add token to cache
	cache.TokenSaved(token)

	// Verify token is in cache
	found, err := cache.FindToken(tokenSecret)
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(100)))

	// Revoke task
	cache.TaskRevoked(taskID)

	// Verify token removed from cache
	_, err = cache.FindToken(tokenSecret)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("token"))
}

// TestTaskRevokedMultipleTokens tests that TaskRevoked removes only tokens for specified task.
func TestTaskRevokedMultipleTokens(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create two tasks
	task1ID := uint(501)
	task2ID := uint(502)

	// Create tokens for each task
	token1Secret := "task1-token"
	token1 := &Token{
		Token: model.Token{
			Model:      Model{ID: 101},
			TaskID:     &task1ID,
			Digest:     secret.Hash(token1Secret),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: token1Secret,
	}

	token2Secret := "task2-token"
	token2 := &Token{
		Token: model.Token{
			Model:      Model{ID: 102},
			TaskID:     &task2ID,
			Digest:     secret.Hash(token2Secret),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: token2Secret,
	}

	// Add tokens to cache
	cache.TokenSaved(token1)
	cache.TokenSaved(token2)

	// Verify both tokens are in cache
	_, err = cache.FindToken(token1Secret)
	g.Expect(err).To(BeNil())
	_, err = cache.FindToken(token2Secret)
	g.Expect(err).To(BeNil())

	// Revoke only task1
	cache.TaskRevoked(task1ID)

	// Verify task1 token removed
	_, err = cache.FindToken(token1Secret)
	g.Expect(err).NotTo(BeNil())

	// Verify task2 token still exists
	found, err := cache.FindToken(token2Secret)
	g.Expect(err).To(BeNil())
	g.Expect(found.ID).To(Equal(uint(102)))
}

// TestTaskRevokedTransaction tests TaskRevoked within a transaction.
func TestTaskRevokedTransaction(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	taskID := uint(503)
	tokenSecret := "tx-task-token"
	token := &Token{
		Token: model.Token{
			Model:      Model{ID: 103},
			TaskID:     &taskID,
			Digest:     secret.Hash(tokenSecret),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: tokenSecret,
	}

	// Add token
	cache.TokenSaved(token)

	// Verify in cache
	_, err = cache.FindToken(tokenSecret)
	g.Expect(err).To(BeNil())

	// Revoke within successful transaction
	err = cache.Transaction(func(tx *Tx) error {
		tx.TaskRevoked(taskID)
		return nil
	})
	g.Expect(err).To(BeNil())

	// Verify token removed
	_, err = cache.FindToken(tokenSecret)
	g.Expect(err).NotTo(BeNil())

	// Test rollback
	task2ID := uint(504)
	token2Secret := "tx-task2-token"
	token2 := &Token{
		Token: model.Token{
			Model:      Model{ID: 104},
			TaskID:     &task2ID,
			Digest:     secret.Hash(token2Secret),
			Expiration: time.Now().Add(24 * time.Hour),
		},
		Secret: token2Secret,
	}

	cache.TokenSaved(token2)

	// Rollback transaction
	err = cache.Transaction(func(tx *Tx) error {
		tx.TaskRevoked(task2ID)
		return fmt.Errorf("rollback test")
	})
	g.Expect(err).NotTo(BeNil())

	// Verify token still exists (rollback successful)
	_, err = cache.FindToken(token2Secret)
	g.Expect(err).To(BeNil())
}

// TestCacheUserSavedBySubject tests that UserSaved updates bySubject map.
func TestCacheUserSavedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create user not in DB (just for cache testing)
	user := &User{
		Model:    Model{ID: 999},
		Subject:  "cached-user-subject",
		Login:    "cacheduser",
		Email:    "cached@example.com",
		Password: secret.HashPassword("password"),
	}

	// Save to cache
	cache.UserSaved(user)

	// Verify it's in both maps
	d := cache.data.Load()
	userById, foundById := d.userById[999]
	userBySubject, foundBySubject := d.userBySubject["cached-user-subject"]

	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())
	g.Expect(userById.Login).To(Equal("cacheduser"))
	g.Expect(userBySubject.Login).To(Equal("cacheduser"))

	// Verify FindSubject works without DB query
	subject, err := cache.FindSubject("cached-user-subject")
	g.Expect(err).To(BeNil())
	g.Expect(subject.Key).To(Equal("cached-user-subject"))
	g.Expect(subject.User.Login).To(Equal("cacheduser"))
}

// TestCacheIdentitySavedBySubject tests that IdentitySaved updates bySubject map.
func TestCacheIdentitySavedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create identity not in DB (just for cache testing)
	identity := &Identity{
		Model:   Model{ID: 888},
		Issuer:  "https://test.idp.com",
		Subject: "cached-identity-subject",
		Login:   "cachedidentity",
		Email:   "cachedidentity@example.com",
		Scopes:  "openid profile",
	}

	// Save to cache
	cache.IdentitySaved(identity)

	// Verify it's in both maps
	d := cache.data.Load()
	identById, foundById := d.identById[888]
	identBySubject, foundBySubject := d.identBySubject["cached-identity-subject"]

	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())
	g.Expect(identById.Login).To(Equal("cachedidentity"))
	g.Expect(identBySubject.Login).To(Equal("cachedidentity"))

	// Verify FindSubject works without DB query
	subject, err := cache.FindSubject("cached-identity-subject")
	g.Expect(err).To(BeNil())
	g.Expect(subject.Key).To(Equal("cached-identity-subject"))
	g.Expect(subject.Identity.Login).To(Equal("cachedidentity"))
}

// TestCacheUserDeletedBySubject tests that UserDeleted removes from bySubject map.
func TestCacheUserDeletedBySubject(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	user := &User{
		Model:   Model{ID: 777},
		Subject: "delete-user-subject",
		Login:   "deleteuser",
	}

	cache.UserSaved(user)

	// Verify it's in both maps
	d := cache.data.Load()
	_, foundById := d.userById[777]
	_, foundBySubject := d.userBySubject["delete-user-subject"]
	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())

	// Delete user
	cache.UserDeleted(777)

	// Verify removed from both maps
	d = cache.data.Load()
	_, foundById = d.userById[777]
	_, foundBySubject = d.userBySubject["delete-user-subject"]
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

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	identity := &Identity{
		Model:   Model{ID: 666},
		Subject: "delete-identity-subject",
		Login:   "deleteidentity",
	}

	cache.IdentitySaved(identity)

	// Verify it's in both maps
	d := cache.data.Load()
	_, foundById := d.identById[666]
	_, foundBySubject := d.identBySubject["delete-identity-subject"]
	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())

	// Delete identity
	cache.IdentityDeleted(666)

	// Verify removed from both maps
	d = cache.data.Load()
	_, foundById = d.identById[666]
	_, foundBySubject = d.identBySubject["delete-identity-subject"]
	g.Expect(foundById).To(BeFalse())
	g.Expect(foundBySubject).To(BeFalse())

	// Verify FindSubject returns NotFound
	_, err = cache.FindSubject("delete-identity-subject")
	g.Expect(err).NotTo(BeNil())
}

// TestCacheUserByUseridMaps tests that user is in all three maps.
func TestCacheUserByUseridMaps(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Create user
	user := &User{
		Model:    Model{ID: 555},
		Subject:  "map-test-subject",
		Login:    "maptestuser",
		Email:    "maptest@example.com",
		Password: secret.HashPassword("password"),
	}

	// Save to cache
	cache.UserSaved(user)

	// Verify it's in all three maps
	d := cache.data.Load()
	userById, foundById := d.userById[555]
	userBySubject, foundBySubject := d.userBySubject["map-test-subject"]
	userByLogin, foundByLogin := d.userByLogin["maptestuser"]

	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())
	g.Expect(foundByLogin).To(BeTrue())
	g.Expect(userById.Login).To(Equal("maptestuser"))
	g.Expect(userBySubject.Login).To(Equal("maptestuser"))
	g.Expect(userByLogin.Subject).To(Equal("map-test-subject"))

	// Delete user
	cache.UserDeleted(555)

	// Verify removed from all three maps
	d = cache.data.Load()
	_, foundById = d.userById[555]
	_, foundBySubject = d.userBySubject["map-test-subject"]
	_, foundByLogin = d.userByLogin["maptestuser"]

	g.Expect(foundById).To(BeFalse())
	g.Expect(foundBySubject).To(BeFalse())
	g.Expect(foundByLogin).To(BeFalse())
}

// TestCacheConcurrency tests concurrent access to the cache.
func TestCacheConcurrency(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create multiple users in DB
	users := make([]*User, 10)
	for i := 0; i < 10; i++ {
		users[i] = &User{
			Subject:  fmt.Sprintf("concurrent-user-%d", i),
			Login:    fmt.Sprintf("concurrentuser%d", i),
			Password: secret.HashPassword("password"),
			Email:    fmt.Sprintf("concurrent%d@example.com", i),
		}
		err = db.Create(users[i]).Error
		g.Expect(err).To(BeNil())
	}

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Launch concurrent FindUserByLogin operations
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userIdx := idx % 10
			_, err := cache.FindUserByLogin(fmt.Sprintf("concurrentuser%d", userIdx))
			g.Expect(err).To(BeNil())
		}(i)
	}

	// Launch concurrent UserSaved operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userIdx := idx % 10
			cache.UserSaved((*User)(users[userIdx]))
		}(i)
	}

	wg.Wait()
}

// TestManualCacheRefresh tests Reset and Refresh operations.
func TestManualCacheRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	user := &User{
		Subject:  "manual-refresh-user",
		Login:    "manualrefreshuser",
		Password: secret.HashPassword("password"),
		Email:    "manualrefresh@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Verify user is in cache
	found, err := cache.FindUserByLogin("manualrefreshuser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())

	// Call Reset - should clear cache
	cache.Reset()

	// Cache is now empty, user not found
	_, err = cache.FindUserByLogin("manualrefreshuser")
	g.Expect(err).NotTo(BeNil())

	// Call manual Refresh to reload
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// User should still be found
	found, err = cache.FindUserByLogin("manualrefreshuser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
}

// TestFindRoleByName tests finding roles by name.
func TestFindRoleByName(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test roles
	perm1 := &Permission{
		Name:  "Read Apps",
		Scope: "applications:get",
	}
	err = db.Create(perm1).Error
	g.Expect(err).To(BeNil())

	perm2 := &Permission{
		Name:  "Write Apps",
		Scope: "applications:post",
	}
	err = db.Create(perm2).Error
	g.Expect(err).To(BeNil())

	role1 := &Role{
		Name:        "AppReader",
		Permissions: []Permission{*perm1},
	}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())

	role2 := &Role{
		Name:        "AppWriter",
		Permissions: []Permission{*perm2},
	}
	err = db.Create(role2).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test finding existing roles by name
	found1, err := cache.FindRoleByName("AppReader")
	g.Expect(err).To(BeNil())
	g.Expect(found1).NotTo(BeNil())
	g.Expect(found1.Name).To(Equal("AppReader"))
	g.Expect(found1.GetScopes()).To(ContainElement("applications:get"))

	found2, err := cache.FindRoleByName("AppWriter")
	g.Expect(err).To(BeNil())
	g.Expect(found2).NotTo(BeNil())
	g.Expect(found2.Name).To(Equal("AppWriter"))
	g.Expect(found2.GetScopes()).To(ContainElement("applications:post"))

	// Test finding non-existent role
	_, err = cache.FindRoleByName("NonExistentRole")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("Role"))
	g.Expect(notFound.Id).To(Equal("NonExistentRole"))
}

// TestFindRoleById tests finding roles by ID.
func TestFindRoleById(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	perm := &Permission{
		Name:  "Read Tasks",
		Scope: "tasks:get",
	}
	err = db.Create(perm).Error
	g.Expect(err).To(BeNil())

	role := &Role{
		Name:        "TaskReader",
		Permissions: []Permission{*perm},
	}
	err = db.Create(role).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test finding existing role by ID
	found, err := cache.FindRoleById(role.ID)
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Name).To(Equal("TaskReader"))
	g.Expect(found.GetScopes()).To(ContainElement("tasks:get"))

	// Test finding non-existent role
	_, err = cache.FindRoleById(9999)
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("Role"))
}

// TestFindIdentityByLogin tests finding identities by userid.
func TestFindIdentityByLogin(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	identity := &Identity{
		Issuer:  "https://idp.example.com",
		Subject: "idp-subject-123",
		Login:   "idpuser123",
		Email:   "idp@example.com",
		Scopes:  "openid profile email",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test finding existing identity by userid
	found, err := cache.FindIdentityByLogin("idpuser123")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(Equal("idp-subject-123"))
	g.Expect(found.Email).To(Equal("idp@example.com"))

	// Test finding non-existent identity
	_, err = cache.FindIdentityByLogin("nonexistent")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("identity"))
}

// TestFindTokenEdgeCases tests edge cases in token finding.
func TestFindTokenEdgeCases(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test finding non-existent token
	_, err = cache.FindToken("nonexistent-token")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("token"))

	// Test token with no bindings (all foreign keys nil)
	unboundToken := &Token{
		Token: model.Token{
			Kind:          KindAPIKey,
			AuthId:        "unbound-auth-id",
			Digest:        secret.Hash("unbound-secret"),
			Expiration:    time.Now().Add(24 * time.Hour),
			UserID:        nil,
			TaskID:        nil,
			IdpIdentityID: nil,
		},
		Secret: "unbound-secret",
	}
	err = db.Create(&unboundToken.Token).Error
	g.Expect(err).To(BeNil())

	cache.TokenSaved(unboundToken)

	found, err := cache.FindToken("unbound-secret")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(BeEmpty())
	g.Expect(found.Scopes).To(BeEmpty())

	// Test expired token is not cached
	expiredToken := &Token{
		Token: model.Token{
			Kind:       KindAPIKey,
			AuthId:     "expired-auth-id",
			Digest:     secret.Hash("expired-secret"),
			Expiration: time.Now().Add(-1 * time.Hour), // Expired
		},
		Secret: "expired-secret",
	}
	err = db.Create(&expiredToken.Token).Error
	g.Expect(err).To(BeNil())

	// Refresh cache - expired tokens should not be loaded
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	_, err = cache.FindToken("expired-secret")
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
}

// TestFindSubjectEdgeCases tests edge cases in subject finding.
func TestFindSubjectEdgeCases(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test finding non-existent subject
	_, err = cache.FindSubject("nonexistent-subject")
	g.Expect(err).NotTo(BeNil())
	var notFound *NotFound
	g.Expect(errors.As(err, &notFound)).To(BeTrue())
	g.Expect(notFound.Resource).To(Equal("subject"))
	g.Expect(notFound.Id).To(Equal("nonexistent-subject"))
}

// TestRoleGetScopes tests Role.GetScopes method.
func TestRoleGetScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	// Test role with multiple permissions
	role := &Role{
		Model: Model{ID: 1},
		Name:  "TestRole",
		Permissions: []Permission{
			{Scope: "applications:get"},
			{Scope: "applications:post"},
			{Scope: "tasks:get"},
		},
	}

	scopes := role.GetScopes()
	g.Expect(scopes).To(HaveLen(3))
	g.Expect(scopes).To(ContainElement("applications:get"))
	g.Expect(scopes).To(ContainElement("applications:post"))
	g.Expect(scopes).To(ContainElement("tasks:get"))

	// Test role with duplicate scopes
	roleWithDupes := &Role{
		Model: Model{ID: 2},
		Name:  "RoleWithDupes",
		Permissions: []Permission{
			{Scope: "tasks:get"},
			{Scope: "applications:get"},
			{Scope: "tasks:get"}, // Duplicate
		},
	}

	scopes = roleWithDupes.GetScopes()
	g.Expect(scopes).To(HaveLen(2)) // Duplicates removed
	g.Expect(scopes).To(ContainElement("tasks:get"))
	g.Expect(scopes).To(ContainElement("applications:get"))

	// Test role with no permissions
	emptyRole := &Role{
		Model:       Model{ID: 3},
		Name:        "EmptyRole",
		Permissions: []Permission{},
	}

	scopes = emptyRole.GetScopes()
	g.Expect(scopes).To(BeEmpty())
}

// TestUserGetScopes tests User.GetScopes method.
func TestUserGetScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create permissions
	perm1 := &Permission{Name: "Perm1", Scope: "apps:get"}
	perm2 := &Permission{Name: "Perm2", Scope: "apps:post"}
	perm3 := &Permission{Name: "Perm3", Scope: "tasks:get"}
	err = db.Create(perm1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(perm2).Error
	g.Expect(err).To(BeNil())
	err = db.Create(perm3).Error
	g.Expect(err).To(BeNil())

	// Create roles
	role1 := &Role{
		Name:        "Role1",
		Permissions: []Permission{*perm1, *perm2},
	}
	role2 := &Role{
		Name:        "Role2",
		Permissions: []Permission{*perm2, *perm3}, // perm2 is duplicate
	}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(role2).Error
	g.Expect(err).To(BeNil())

	// Create user with both roles
	user := &User{
		Subject: "test-user",
		Login:   "testuser",
		Roles:   []model.Role{model.Role(*role1), model.Role(*role2)},
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	cachedUser, err := cache.FindUserByLogin("testuser")
	g.Expect(err).To(BeNil())

	scopes := cachedUser.GetScopes(cache)
	g.Expect(scopes).To(HaveLen(3)) // Should have 3 unique scopes
	g.Expect(scopes).To(ContainElement("apps:get"))
	g.Expect(scopes).To(ContainElement("apps:post"))
	g.Expect(scopes).To(ContainElement("tasks:get"))

	// Test user with no roles
	userNoRoles := &User{
		Model:   Model{ID: 999},
		Subject: "noroles",
		Login:   "noroles",
	}
	scopes = userNoRoles.GetScopes(cache)
	g.Expect(scopes).To(BeEmpty())
}

// TestIdpClientWith tests IdpClient.With() method.
func TestIdpClientWith(t *testing.T) {
	g := NewGomegaWithT(t)

	// Test with all fields populated
	settingsClient := &as.IdpClient{
		ID:              1,
		ClientId:        "test-client",
		Secret:          "test-secret",
		ApplicationType: "web",
		Grants:          []string{"authorization_code", "refresh_token"},
		RedirectURIs:    []string{"http://localhost/callback"},
		Scopes:          []string{"openid", "profile"},
	}

	cacheClient := &IdpClient{}
	cacheClient.With(settingsClient)

	g.Expect(cacheClient.ID).To(Equal(uint(1)))
	g.Expect(cacheClient.ClientId).To(Equal("test-client"))
	g.Expect(cacheClient.Secret).To(Equal("test-secret"))
	g.Expect(cacheClient.ApplicationType).To(Equal("web"))
	g.Expect(cacheClient.Grants).To(HaveLen(2))
	g.Expect(cacheClient.Grants).To(ContainElement("authorization_code"))
	g.Expect(cacheClient.Grants).To(ContainElement("refresh_token"))
	g.Expect(cacheClient.RedirectURIs).To(HaveLen(1))
	g.Expect(cacheClient.RedirectURIs[0]).To(Equal("http://localhost/callback"))
	g.Expect(cacheClient.Scopes).To(HaveLen(2))
	g.Expect(cacheClient.Scopes).To(ContainElement("openid"))
	g.Expect(cacheClient.Scopes).To(ContainElement("profile"))

	// Test with empty secret (public client)
	publicClient := &as.IdpClient{
		ID:              2,
		ClientId:        "public-client",
		Secret:          "",
		ApplicationType: "native",
		Grants:          []string{"device_code"},
		Scopes:          []string{"openid"},
	}

	cachePublic := &IdpClient{}
	cachePublic.With(publicClient)

	g.Expect(cachePublic.ID).To(Equal(uint(2)))
	g.Expect(cachePublic.ClientId).To(Equal("public-client"))
	g.Expect(cachePublic.Secret).To(BeEmpty())
	g.Expect(cachePublic.ApplicationType).To(Equal("native"))
}

// TestCopyOnWriteCommit verifies Commit clones Data instead of mutating.
func TestCopyOnWriteCommit(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Load Data before Tx
	oldData := cache.data.Load()

	// Execute transaction
	user := &User{
		Model:   Model{ID: 100},
		Subject: "cow-test",
		Login:   "cowtest",
	}
	cache.UserSaved(user)

	// Load Data after Tx
	newData := cache.data.Load()

	// Verify Data pointer changed (copy-on-write)
	g.Expect(oldData).NotTo(Equal(newData))

	// Verify old Data unchanged
	_, found := oldData.userById[100]
	g.Expect(found).To(BeFalse())

	// Verify new Data has change
	_, found = newData.userById[100]
	g.Expect(found).To(BeTrue())
}

// TestConcurrentCommitsNoLostUpdates verifies mutex prevents lost updates.
func TestConcurrentCommitsNoLostUpdates(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Spawn 10 concurrent UserSaved operations
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user := &User{
				Model:   Model{ID: uint(id + 100)},
				Subject: fmt.Sprintf("concurrent-%d", id),
				Login:   fmt.Sprintf("concurrent%d", id),
			}
			cache.UserSaved(user)
		}(i)
	}

	wg.Wait()

	// Verify ALL 10 users are in cache (no lost updates)
	d := cache.data.Load()
	for i := 0; i < 10; i++ {
		_, found := d.userById[uint(i+100)]
		g.Expect(found).To(BeTrue(), fmt.Sprintf("user %d not found", i))
	}
}

// TestDataClone verifies clone creates independent copy.
func TestDataClone(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create original Data with content
	original := &Data{}
	original.reset()
	original.userById[1] = &User{Model: Model{ID: 1}, Login: "user1"}
	original.roleById[2] = &Role{Model: Model{ID: 2}, Name: "role1"}
	original.refreshed = time.Now().Add(-5 * time.Minute)

	// Clone it
	clone := original.clone()

	// Verify refreshed timestamp copied
	g.Expect(clone.refreshed).To(Equal(original.refreshed))

	// Verify maps copied (share same objects)
	g.Expect(clone.userById[1]).To(Equal(original.userById[1]))
	g.Expect(clone.roleById[2]).To(Equal(original.roleById[2]))

	// Mutate clone - should NOT affect original
	clone.userById[3] = &User{Model: Model{ID: 3}, Login: "user3"}
	delete(clone.roleById, 2)

	// Verify original unchanged
	_, found := original.userById[3]
	g.Expect(found).To(BeFalse())
	_, found = original.roleById[2]
	g.Expect(found).To(BeTrue())

	// Verify clone has mutations
	_, found = clone.userById[3]
	g.Expect(found).To(BeTrue())
	_, found = clone.roleById[2]
	g.Expect(found).To(BeFalse())
}

// TestOnDemandRefresh verifies ensureRefreshed triggers when stale.
func TestOnDemandRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Get original Data and manually age it
	d := cache.data.Load()
	oldRefreshed := d.refreshed

	// Create a new stale Data and swap it in
	staleData := d.clone()
	staleData.refreshed = time.Now().Add(-10 * time.Minute) // Stale
	cache.data.Store(staleData)

	// Create new user in DB (bypassing cache notifications)
	user := &User{
		Subject:  "on-demand-user",
		Login:    "ondemanduser",
		Password: secret.HashPassword("password"),
		Email:    "ondemand@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	// First FindXXX should trigger refresh
	cache.FindUserByLogin("ondemanduser")

	// Poll until refresh completes (max 2 seconds)
	refreshed := false
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		d = cache.data.Load()
		if d.refreshed.After(oldRefreshed) {
			refreshed = true
			break
		}
	}
	g.Expect(refreshed).To(BeTrue(), "refresh should complete within 2 seconds")

	// Verify new user now in cache
	found, err := cache.FindUserByLogin("ondemanduser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
}

// TestAuthStorm tests lock-free reads under concurrent load.
func TestAuthStorm(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test tokens
	for i := 0; i < 10; i++ {
		token := &Token{
			Token: model.Token{
				Kind:       KindAPIKey,
				AuthId:     fmt.Sprintf("storm-auth-%d", i),
				Digest:     secret.Hash(fmt.Sprintf("storm-token-%d", i)),
				Expiration: time.Now().Add(24 * time.Hour),
			},
			Secret: fmt.Sprintf("storm-token-%d", i),
		}
		err = db.Create(&token.Token).Error
		g.Expect(err).To(BeNil())
	}

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// 100 goroutines each doing 10 FindToken calls
	var wg sync.WaitGroup
	successCount := make([]int, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				tokenIdx := j % 10
				_, err := cache.FindToken(fmt.Sprintf("storm-token-%d", tokenIdx))
				if err == nil {
					successCount[idx]++
				}
			}
		}(i)
	}

	wg.Wait()

	// All reads should succeed
	totalSuccess := 0
	for _, count := range successCount {
		totalSuccess += count
	}
	g.Expect(totalSuccess).To(Equal(1000))
}

// TestTaskChurnUnderLoad tests copy-on-write under write load.
func TestTaskChurnUnderLoad(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	var wg sync.WaitGroup
	doneChan := make(chan struct{})

	// 10 writers: create/revoke task tokens
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskID := uint(1000 + id)
			tokenSecret := fmt.Sprintf("churn-token-%d", id)

			for j := 0; j < 10; j++ {
				select {
				case <-doneChan:
					return
				default:
				}

				// Add token
				token := &Token{
					Token: model.Token{
						Model:      Model{ID: uint(id*100 + j)},
						TaskID:     &taskID,
						Digest:     secret.Hash(tokenSecret),
						Expiration: time.Now().Add(24 * time.Hour),
					},
					Secret: tokenSecret,
				}
				cache.TokenSaved(token)

				// Revoke it
				cache.TaskRevoked(taskID)
			}
		}(i)
	}

	// 50 readers: continuously try to find tokens
	readErrors := make([]int, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				select {
				case <-doneChan:
					return
				default:
				}
				tokenIdx := j % 10
				_, err := cache.FindToken(fmt.Sprintf("churn-token-%d", tokenIdx))
				// Token may or may not exist (being churned)
				// Just verify no panic/crash
				if err != nil {
					readErrors[idx]++
				}
			}
		}(i)
	}

	wg.Wait()
	close(doneChan)

	// Test passes if no panics occurred
	// Read errors are expected due to churn
}

// TestRefreshRace tests that Data.refreshOnce prevents stampede.
func TestRefreshRace(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Make cache stale
	d := cache.data.Load()
	staleData := d.clone()
	staleData.refreshed = time.Now().Add(-10 * time.Minute)
	cache.data.Store(staleData)

	// 50 goroutines all detect staleness simultaneously
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// This triggers ensureRefreshed
			cache.FindUserByLogin("nonexistent")
		}()
	}

	wg.Wait()

	// Give refresh time to complete
	time.Sleep(100 * time.Millisecond)

	// Verify cache refreshed (exactly once via sync.Once)
	d = cache.data.Load()
	g.Expect(d.refreshed).To(BeTemporally(">", staleData.refreshed))
}

// TestMixedConcurrentOperations tests all operations work concurrently.
func TestMixedConcurrentOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	var wg sync.WaitGroup
	iterations := 20

	// Concurrent UserSaved
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				user := &User{
					Model:   Model{ID: uint(id*1000 + j)},
					Subject: fmt.Sprintf("mixed-user-%d-%d", id, j),
					Login:   fmt.Sprintf("mixeduser%d-%d", id, j),
				}
				cache.UserSaved(user)
			}
		}(i)
	}

	// Concurrent RoleSaved/Deleted
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
				cache.FindUserByLogin(fmt.Sprintf("mixeduser%d-%d", userIdx, j))
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	d := cache.data.Load()
	// Should have some users (not all, some deleted/not found)
	g.Expect(len(d.userById)).To(BeNumerically(">", 0))
}
