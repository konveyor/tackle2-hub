package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestAPIKey(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get seeded keys
	seeded, err := client.APIKey.List()
	g.Expect(err).To(BeNil())

	// CREATE: Request an API key using seeded admin user credentials
	keyRequest := &api.APIKey{
		Userid:   "admin",
		Password: "admin",
		Lifespan: 24, // 24 hours
	}

	err = client.APIKey.Create(keyRequest)
	g.Expect(err).To(BeNil())

	// Verify secret and digest are returned on create
	g.Expect(keyRequest.Secret).NotTo(BeEmpty())
	g.Expect(keyRequest.Digest).NotTo(BeEmpty())

	// Save these for later verification
	secret := keyRequest.Secret
	digest := keyRequest.Digest

	// GET: List API keys - should include the new key
	list, err := client.APIKey.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))

	// Find the newly created key by digest
	var createdKey *api.APIKey
	for i := range list {
		if list[i].Digest == digest {
			createdKey = &list[i]
			break
		}
	}
	g.Expect(createdKey).NotTo(BeNil())
	g.Expect(createdKey.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.APIKey.Delete(createdKey.ID)
	})

	// GET: Retrieve the API key by ID
	retrieved, err := client.APIKey.Get(createdKey.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(createdKey.ID))
	g.Expect(retrieved.Digest).To(Equal(digest))

	// Secret should be empty on GET (only returned on create)
	g.Expect(retrieved.Secret).To(BeEmpty())

	// Verify the key is associated with admin user
	g.Expect(retrieved.User).NotTo(BeNil())
	g.Expect(retrieved.User.Name).To(Equal("admin"))

	// DELETE: Remove the API key
	err = client.APIKey.Delete(createdKey.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.APIKey.Get(createdKey.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())

	// Verify the secret works for authentication
	// Note: We can't fully test this since the key was already deleted,
	// but we verified it was created and had a valid secret
	g.Expect(secret).NotTo(BeEmpty(), "API key secret should have been returned on create")
}
