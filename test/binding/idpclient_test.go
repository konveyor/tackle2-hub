package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestIdpClient(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get seeded clients
	seeded, err := client.IdpClient.List()
	g.Expect(err).To(BeNil())

	// Define the client to create
	idpClient := &api.IdpClient{
		ClientId:        "test-client",
		Secret:          "test-secret",
		ApplicationType: "native",
		Grants: []string{
			"authorization_code",
			"refresh_token",
		},
		RedirectURIs: []string{
			"http://localhost:8080/callback",
		},
		Scopes: []string{
			"openid",
			"profile",
			"email",
		},
	}

	// CREATE: Create the client
	err = client.IdpClient.Create(idpClient)
	g.Expect(err).To(BeNil())
	g.Expect(idpClient.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.IdpClient.Delete(idpClient.ID)
	})

	// GET: List clients
	list, err := client.IdpClient.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))

	// GET: Retrieve the client and verify it matches
	retrieved, err := client.IdpClient.Get(idpClient.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(idpClient, retrieved, "Secret")
	g.Expect(eq).To(BeTrue(), report)

	// Verify secret is not exposed in GET
	g.Expect(retrieved.Secret).To(Equal(api.SecretMask))

	// Verify other fields
	g.Expect(retrieved.ClientId).To(Equal("test-client"))
	g.Expect(retrieved.ApplicationType).To(Equal("native"))
	g.Expect(len(retrieved.Grants)).To(Equal(2))
	g.Expect(retrieved.Grants).To(ContainElement("authorization_code"))
	g.Expect(retrieved.Grants).To(ContainElement("refresh_token"))
	g.Expect(len(retrieved.RedirectURIs)).To(Equal(1))
	g.Expect(retrieved.RedirectURIs[0]).To(Equal("http://localhost:8080/callback"))
	g.Expect(len(retrieved.Scopes)).To(Equal(3))

	// UPDATE: Modify the client
	idpClient.ApplicationType = "web"
	idpClient.Grants = []string{
		"authorization_code",
		"refresh_token",
		"urn:ietf:params:oauth:grant-type:jwt-bearer",
	}
	idpClient.RedirectURIs = []string{
		"http://localhost:3000/callback",
		"http://localhost:3000/silent-callback",
	}

	err = client.IdpClient.Update(idpClient)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.IdpClient.Get(idpClient.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(idpClient, updated, "UpdateUser", "Secret")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the client
	err = client.IdpClient.Delete(idpClient.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.IdpClient.Get(idpClient.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestIdpClientProtectedIDs tests that seeded clients (ID < 1000) cannot be modified or deleted.
func TestIdpClientProtectedIDs(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get seeded clients
	seeded, err := client.IdpClient.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(seeded)).To(BeNumerically(">=", 1))

	// Find a seeded client (ID < 1000)
	var seededClient *api.IdpClient
	for i := range seeded {
		if seeded[i].ID < 1000 {
			seededClient = &seeded[i]
			break
		}
	}
	g.Expect(seededClient).NotTo(BeNil(), "Should have at least one seeded client")

	// Attempt to update seeded client (should fail)
	seededClient.ApplicationType = "updated"
	err = client.IdpClient.Update(seededClient)
	g.Expect(err).NotTo(BeNil(), "Should not be able to update seeded client")

	// Attempt to delete seeded client (should fail)
	err = client.IdpClient.Delete(seededClient.ID)
	g.Expect(err).NotTo(BeNil(), "Should not be able to delete seeded client")

	// Verify seeded client still exists unchanged
	unchanged, err := client.IdpClient.Get(seededClient.ID)
	g.Expect(err).To(BeNil())
	g.Expect(unchanged).NotTo(BeNil())
	g.Expect(unchanged.ApplicationType).NotTo(Equal("updated"))
}

// TestIdpClientCreate tests client creation with ID reservation.
func TestIdpClientCreateWithReservedID(t *testing.T) {
	g := NewGomegaWithT(t)

	// Attempt to create client with reserved ID (< 1000)
	idpClient := &api.IdpClient{
		ClientId:        "reserved-id-test",
		ApplicationType: "native",
		Grants:          []string{"authorization_code"},
		Scopes:          []string{"openid"},
	}
	idpClient.ID = 500 // Reserved ID

	err := client.IdpClient.Create(idpClient)
	g.Expect(err).NotTo(BeNil(), "Should not be able to create client with reserved ID")
}
