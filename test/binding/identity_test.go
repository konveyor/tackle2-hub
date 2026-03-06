package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestIdentity(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the identity to create
	password := "test-password-123"
	key := "test-key-123"
	settings := "{\"insecureSkipVerify\": true}"
	identity := &api.Identity{
		Name:        "test-git-identity",
		Kind:        "git",
		Description: "Git identity for testing",
		User:        "test-user",
		Password:    password,
		Key:         key,
		Settings:    settings,
		Default:     false,
	}

	// CREATE: Create the identity
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	g.Expect(identity.ID).NotTo(BeZero())
	g.Expect(identity.Password).ToNot(Equal(password)) // encrypted.
	g.Expect(identity.Key).ToNot(Equal(key))           // encrypted.
	g.Expect(identity.Settings).ToNot(Equal(settings)) // encrypted.
	identity.Password = password
	identity.Key = key
	identity.Settings = settings
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// GET: List identities
	list, err := client.Identity.Decrypted().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(identity, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the identity and verify it matches
	retrieved, err := client.Identity.Decrypted().Get(identity.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(identity, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the identity
	identity.Name = "updated-git-identity"
	identity.User = "updated-user"
	identity.Password = "updated-password-456"
	identity.Description = "Updated description"

	err = client.Identity.Update(identity)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Identity.Decrypted().Get(identity.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(identity, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the identity
	err = client.Identity.Delete(identity.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Identity.Get(identity.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestIdentityDecryption tests the Decrypt() method for encrypted fields
func TestIdentityDecryption(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the identity with secrets
	password := "test-password-123"
	key := "test-key-123"
	settings := "{\"insecureSkipVerify\": true}"
	identity := &api.Identity{
		Name:        "test-decrypt-identity",
		Kind:        "git",
		Description: "Identity for testing decryption",
		User:        "test-user",
		Password:    password,
		Key:         key,
		Settings:    settings,
		Default:     false,
	}

	// CREATE: Create the identity
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	g.Expect(identity.ID).NotTo(BeZero())
	g.Expect(identity.Password).ToNot(Equal(password)) // encrypted.
	g.Expect(identity.Key).ToNot(Equal(key))           // encrypted.
	g.Expect(identity.Settings).ToNot(Equal(settings)) // encrypted.
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// Build expected identity with decrypted values
	expected := &api.Identity{
		Name:        "test-decrypt-identity",
		Kind:        "git",
		Description: "Identity for testing decryption",
		User:        "test-user",
		Password:    password,
		Key:         key,
		Settings:    settings,
		Default:     false,
	}
	expected.ID = identity.ID

	// GET without Decrypt - verify fields are encrypted
	encrypted, err := client.Identity.Get(identity.ID)
	g.Expect(err).To(BeNil())
	g.Expect(encrypted).NotTo(BeNil())
	g.Expect(encrypted.Password).ToNot(Equal(password))
	g.Expect(encrypted.Key).ToNot(Equal(key))
	g.Expect(encrypted.Settings).ToNot(Equal(settings))

	// GET with Decrypt - verify fields are decrypted
	decrypted, err := client.Identity.Decrypted().Get(identity.ID)
	g.Expect(err).To(BeNil())
	g.Expect(decrypted).NotTo(BeNil())
	eq, report := cmp.Eq(expected, decrypted, "CreateUser", "UpdateUser", "CreateTime")
	g.Expect(eq).To(BeTrue(), report)

	// LIST without Decrypt - verify fields are encrypted
	encryptedList, err := client.Identity.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(encryptedList)).To(BeNumerically(">", 0))
	found := false
	for _, id := range encryptedList {
		if id.ID == identity.ID {
			found = true
			g.Expect(id.Password).ToNot(Equal(password))
			g.Expect(id.Key).ToNot(Equal(key))
			g.Expect(id.Settings).ToNot(Equal(settings))
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// LIST with Decrypt - verify fields are decrypted
	decryptedList, err := client.Identity.Decrypted().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(decryptedList)).To(BeNumerically(">", 0))
	found = false
	for _, id := range decryptedList {
		if id.ID == identity.ID {
			found = true
			eq, report = cmp.Eq(expected, &id, "CreateUser", "UpdateUser", "CreateTime")
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// FIND without Decrypt - verify fields are encrypted
	filter := binding.Filter{}
	filter.And("name").Eq(identity.Name)
	encryptedFound, err := client.Identity.Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(encryptedFound)).To(BeNumerically(">", 0))
	found = false
	for _, id := range encryptedFound {
		if id.ID == identity.ID {
			found = true
			g.Expect(id.Password).ToNot(Equal(password))
			g.Expect(id.Key).ToNot(Equal(key))
			g.Expect(id.Settings).ToNot(Equal(settings))
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// FIND with Decrypt - verify fields are decrypted
	decryptedFound, err := client.Identity.Decrypted().Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(decryptedFound)).To(BeNumerically(">", 0))
	found = false
	for _, id := range decryptedFound {
		if id.ID == identity.ID {
			found = true
			eq, report = cmp.Eq(expected, &id, "CreateUser", "UpdateUser", "CreateTime")
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())
}

// TestIdentityFind tests finding identities using filter
func TestIdentityFind(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create first identity
	direct := &api.Identity{
		Name: "direct",
		Kind: "Test",
	}
	err := client.Identity.Create(direct)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(direct.ID)
	})

	// Create second identity with different kind
	direct2 := &api.Identity{
		Name: "direct2",
		Kind: "Other",
	}
	err = client.Identity.Create(direct2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(direct2.ID)
	})

	// Create application with first identity
	application := &api.Application{
		Name:       "Test App for Identity Find",
		Identities: []api.IdentityRef{{ID: direct.ID}},
	}
	err = client.Application.Create(application)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

	// FIND: Find identity using filter for application.id and kind
	filter := binding.Filter{}
	filter.And("application.id").Eq(int(application.ID))
	filter.And("kind").Eq(direct.Kind)
	found, err := client.Identity.Find(filter)
	g.Expect(err).To(BeNil())
	g.Expect(len(found)).To(BeNumerically(">", 0), "Should find at least one identity")

	// Verify found identity is the correct one
	identity := found[0]
	g.Expect(identity.ID).To(Equal(direct.ID))
	g.Expect(identity.Kind).To(Equal(direct.Kind))
}
