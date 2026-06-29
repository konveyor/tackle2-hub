package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestToken(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a PAT token
	req := &api.PAT{}
	err := client.Token.Create(req)
	g.Expect(err).To(BeNil())
	g.Expect(req.ID).NotTo(BeZero())
	g.Expect(req.Token).NotTo(BeEmpty())
	t.Cleanup(func() {
		_ = client.Token.Delete(req.ID)
	})

	// GET: List tokens
	list, err := client.Token.List()
	g.Expect(err).To(BeNil())

	// Verify we have at least one token (from the test client login)
	g.Expect(len(list)).To(BeNumerically(">", 0))

	// GET: Retrieve the created token by ID
	retrieved, err := client.Token.Get(req.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(req.ID))
	g.Expect(retrieved.Kind).To(Equal(api.TokenKindAPIKey))

	// DELETE: Remove the token
	err = client.Token.Delete(req.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Token.Get(req.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

func TestTokenRevoke(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a PAT (API key) token to be revoked
	req := &api.PAT{}
	err := client.Token.Create(req)
	g.Expect(err).To(BeNil())
	g.Expect(req.ID).NotTo(BeZero())
	g.Expect(req.Token).NotTo(BeEmpty())
	t.Cleanup(func() {
		_ = client.Token.Delete(req.ID)
	})

	// GET: Retrieve the created token
	token, err := client.Token.Get(req.ID)
	g.Expect(err).To(BeNil())
	g.Expect(token.ID).To(Equal(req.ID))

	// Verify PAT has no grant
	g.Expect(token.Grant).To(BeNil())

	// REVOKE: Revoke the PAT token
	err = client.Token.Revoke(req.ID)
	g.Expect(err).To(BeNil())

	// Verify token was deleted by revocation
	_, err = client.Token.Get(req.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// Note: Revocation of grant-backed tokens (created via OIDC authorization flow)
// is tested at the handler/auth layer where prerequisites can be seeded.
// Binding tests focus on PAT revocation which can be tested deterministically.
