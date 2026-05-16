package cache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB() (db *gorm.DB, err error) {
	db, err = gorm.Open(
		sqlite.Open("file::memory:?_foreign_keys=yes"),
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
		Model: model.Model{ID: 100},
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
		Model:   model.Model{ID: 200},
		Subject: "test-user",
		Userid:  "testuser",
	}
	cache.UserSaved(user)
	_, err = cache.FindUserByUserid("testuser")
	g.Expect(err).To(BeNil())

	cache.UserDeleted(200)
	_, err = cache.FindUserByUserid("testuser")
	g.Expect(err).NotTo(BeNil())

	// Test TaskSaved/TaskDeleted
	task := &Task{
		Model: model.Model{ID: 300},
		Name:  "test-task",
		State: "Running",
	}
	cache.TaskSaved(task)
	_, err = cache.FindTaskById(300)
	g.Expect(err).To(BeNil())

	cache.TaskDeleted(300)
	_, err = cache.FindTaskById(300)
	g.Expect(err).NotTo(BeNil())

	// Test IdentitySaved/IdentityDeleted
	identity := &Identity{
		Model:   model.Model{ID: 400},
		Issuer:  "https://idp.example.com",
		Subject: "idp-subject",
		Userid:  "idp-userid",
	}
	cache.IdentitySaved(identity)
	_, err = cache.FindIdentityByUserid("idp-userid")
	g.Expect(err).To(BeNil())

	cache.IdentityDeleted(400)
	_, err = cache.FindIdentityByUserid("idp-userid")
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
		Model: model.Model{ID: 101},
		Name:  "TxRole1",
	}
	role2 := &Role{
		Model: model.Model{ID: 102},
		Name:  "TxRole2",
	}

	err = cache.Transaction(func(tx *Tx) error {
		tx.RoleSaved(role1)
		tx.RoleSaved(role2)
		return nil
	})
	g.Expect(err).To(BeNil())

	// Both roles should be in cache
	cache.mutex.RLock()
	_, found1 := cache.roleById[101]
	_, found2 := cache.roleById[102]
	cache.mutex.RUnlock()
	g.Expect(found1).To(BeTrue())
	g.Expect(found2).To(BeTrue())

	// Test rollback on error
	role3 := &Role{
		Model: model.Model{ID: 103},
		Name:  "TxRole3",
	}

	err = cache.Transaction(func(tx *Tx) error {
		tx.RoleSaved(role3)
		return fmt.Errorf("simulated error")
	})
	g.Expect(err).NotTo(BeNil())

	// Role3 should NOT be in cache (rolled back)
	cache.mutex.RLock()
	_, found3 := cache.roleById[103]
	cache.mutex.RUnlock()
	g.Expect(found3).To(BeFalse())

	// Test explicit Begin/Commit/Rollback
	tx := cache.Begin()
	user := &User{
		Model:   model.Model{ID: 201},
		Subject: "tx-user",
		Userid:  "txuser",
	}
	tx.UserSaved(user)
	tx.Commit()

	cache.mutex.RLock()
	_, foundUser := cache.userById[201]
	cache.mutex.RUnlock()
	g.Expect(foundUser).To(BeTrue())

	// Test rollback
	tx = cache.Begin()
	user2 := &User{
		Model:   model.Model{ID: 202},
		Subject: "tx-user2",
		Userid:  "txuser2",
	}
	tx.UserSaved(user2)
	tx.Rollback() // Discard changes

	cache.mutex.RLock()
	_, foundUser2 := cache.userById[202]
	cache.mutex.RUnlock()
	g.Expect(foundUser2).To(BeFalse())
}

// TestCacheDoubleCheckRefresh tests that concurrent ensureFresh calls don't cause issues.
func TestCacheDoubleCheckRefresh(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test data
	user := &model.User{
		Subject:  "double-check-user",
		Userid:   "doublecheckuser",
		Password: secret.HashPassword("password"),
		Email:    "doublecheck@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Force cache to be stale
	cache.mutex.Lock()
	cache.refreshed = time.Now().Add(-10 * time.Minute)
	cache.mutex.Unlock()

	// Launch multiple concurrent calls to FindUserByUserid (which calls ensureFresh)
	var wg sync.WaitGroup
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := cache.FindUserByUserid("doublecheckuser")
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// All calls should succeed without error
	for i, err := range errors {
		g.Expect(err).To(BeNil(), fmt.Sprintf("goroutine %d failed", i))
	}

	// Cache should have refreshed and found the user
	found, err := cache.FindUserByUserid("doublecheckuser")
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
	g.Expect(subject.Key).To(Equal("cached-user-subject"))
	g.Expect(subject.User.Userid).To(Equal("cacheduser"))
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
	g.Expect(subject.Key).To(Equal("cached-identity-subject"))
	g.Expect(subject.Identity.Userid).To(Equal("cachedidentity"))
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

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

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
		Model:    model.Model{ID: 555},
		Subject:  "map-test-subject",
		Userid:   "maptestuser",
		Email:    "maptest@example.com",
		Password: secret.HashPassword("password"),
	}

	// Save to cache
	cache.UserSaved(user)

	// Verify it's in all three maps
	cache.mutex.RLock()
	userById, foundById := cache.userById[555]
	userBySubject, foundBySubject := cache.userBySubject["map-test-subject"]
	userByUserid, foundByUserid := cache.userByUserid["maptestuser"]
	cache.mutex.RUnlock()

	g.Expect(foundById).To(BeTrue())
	g.Expect(foundBySubject).To(BeTrue())
	g.Expect(foundByUserid).To(BeTrue())
	g.Expect(userById.Userid).To(Equal("maptestuser"))
	g.Expect(userBySubject.Userid).To(Equal("maptestuser"))
	g.Expect(userByUserid.Subject).To(Equal("map-test-subject"))

	// Delete user
	cache.UserDeleted(555)

	// Verify removed from all three maps
	cache.mutex.RLock()
	_, foundById = cache.userById[555]
	_, foundBySubject = cache.userBySubject["map-test-subject"]
	_, foundByUserid = cache.userByUserid["maptestuser"]
	cache.mutex.RUnlock()

	g.Expect(foundById).To(BeFalse())
	g.Expect(foundBySubject).To(BeFalse())
	g.Expect(foundByUserid).To(BeFalse())
}

// TestCacheConcurrency tests concurrent access to the cache.
func TestCacheConcurrency(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create multiple users in DB
	users := make([]*model.User, 10)
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

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Launch concurrent FindUserByUserid operations
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userIdx := idx % 10
			_, err := cache.FindUserByUserid(fmt.Sprintf("concurrentuser%d", userIdx))
			g.Expect(err).To(BeNil())
		}(i)
	}

	// Launch concurrent Refresh operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cache.Refresh()
			g.Expect(err).To(BeNil())
		}()
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

	user := &model.User{
		Subject:  "manual-refresh-user",
		Userid:   "manualrefreshuser",
		Password: secret.HashPassword("password"),
		Email:    "manualrefresh@example.com",
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Verify user is in cache
	found, err := cache.FindUserByUserid("manualrefreshuser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())

	// Call Reset - should clear cache
	cache.Reset()

	// User should still be found via automatic refresh (ensureFresh)
	found, err = cache.FindUserByUserid("manualrefreshuser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())

	// Call manual Refresh
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// User should still be found
	found, err = cache.FindUserByUserid("manualrefreshuser")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
}

// TestFindRoleByName tests finding roles by name.
func TestFindRoleByName(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	// Create test roles
	perm1 := &model.Permission{
		Name:  "Read Apps",
		Scope: "applications:get",
	}
	err = db.Create(perm1).Error
	g.Expect(err).To(BeNil())

	perm2 := &model.Permission{
		Name:  "Write Apps",
		Scope: "applications:post",
	}
	err = db.Create(perm2).Error
	g.Expect(err).To(BeNil())

	role1 := &model.Role{
		Name:        "AppReader",
		Permissions: []model.Permission{*perm1},
	}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())

	role2 := &model.Role{
		Name:        "AppWriter",
		Permissions: []model.Permission{*perm2},
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

	perm := &model.Permission{
		Name:  "Read Tasks",
		Scope: "tasks:get",
	}
	err = db.Create(perm).Error
	g.Expect(err).To(BeNil())

	role := &model.Role{
		Name:        "TaskReader",
		Permissions: []model.Permission{*perm},
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

// TestFindIdentityByUserid tests finding identities by userid.
func TestFindIdentityByUserid(t *testing.T) {
	g := NewGomegaWithT(t)

	db, err := setupTestDB()
	g.Expect(err).To(BeNil())

	identity := &Identity{
		Issuer:  "https://idp.example.com",
		Subject: "idp-subject-123",
		Userid:  "idpuser123",
		Email:   "idp@example.com",
		Scopes:  "openid profile email",
	}
	err = db.Create(identity).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	// Test finding existing identity by userid
	found, err := cache.FindIdentityByUserid("idpuser123")
	g.Expect(err).To(BeNil())
	g.Expect(found).NotTo(BeNil())
	g.Expect(found.Subject).To(Equal("idp-subject-123"))
	g.Expect(found.Email).To(Equal("idp@example.com"))

	// Test finding non-existent identity
	_, err = cache.FindIdentityByUserid("nonexistent")
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
		Model: model.Model{ID: 1},
		Name:  "TestRole",
		Permissions: []model.Permission{
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
		Model: model.Model{ID: 2},
		Name:  "RoleWithDupes",
		Permissions: []model.Permission{
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
		Model:       model.Model{ID: 3},
		Name:        "EmptyRole",
		Permissions: []model.Permission{},
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
	perm1 := &model.Permission{Name: "Perm1", Scope: "apps:get"}
	perm2 := &model.Permission{Name: "Perm2", Scope: "apps:post"}
	perm3 := &model.Permission{Name: "Perm3", Scope: "tasks:get"}
	err = db.Create(perm1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(perm2).Error
	g.Expect(err).To(BeNil())
	err = db.Create(perm3).Error
	g.Expect(err).To(BeNil())

	// Create roles
	role1 := &model.Role{
		Name:        "Role1",
		Permissions: []model.Permission{*perm1, *perm2},
	}
	role2 := &model.Role{
		Name:        "Role2",
		Permissions: []model.Permission{*perm2, *perm3}, // perm2 is duplicate
	}
	err = db.Create(role1).Error
	g.Expect(err).To(BeNil())
	err = db.Create(role2).Error
	g.Expect(err).To(BeNil())

	// Create user with both roles
	user := &model.User{
		Subject: "test-user",
		Userid:  "testuser",
		Roles:   []model.Role{*role1, *role2},
	}
	err = db.Create(user).Error
	g.Expect(err).To(BeNil())

	cache := New(db)
	err = cache.Refresh()
	g.Expect(err).To(BeNil())

	cachedUser, err := cache.FindUserByUserid("testuser")
	g.Expect(err).To(BeNil())

	scopes := cachedUser.GetScopes(cache)
	g.Expect(scopes).To(HaveLen(3)) // Should have 3 unique scopes
	g.Expect(scopes).To(ContainElement("apps:get"))
	g.Expect(scopes).To(ContainElement("apps:post"))
	g.Expect(scopes).To(ContainElement("tasks:get"))

	// Test user with no roles
	userNoRoles := &User{
		Model:   model.Model{ID: 999},
		Subject: "noroles",
		Userid:  "noroles",
	}
	scopes = userNoRoles.GetScopes(cache)
	g.Expect(scopes).To(BeEmpty())
}
