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

func TestTokenRevokePAT(t *testing.T) {
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

func TestTokenRevokeWithGrant(t *testing.T) {
	g := NewGomegaWithT(t)

	// GET: List tokens to find one with a grant
	// Note: Tokens with grants are created via OIDC login, not PAT creation
	tokens, err := client.Token.List()
	g.Expect(err).To(BeNil())

	// Find a token that has an associated grant
	var tokenWithGrant *api.Token
	for i := range tokens {
		if tokens[i].Grant != nil && tokens[i].Grant.ID != 0 {
			tokenWithGrant = &tokens[i]
			break
		}
	}

	if tokenWithGrant == nil {
		t.Skip("No tokens with grants available - requires OIDC login")
		return
	}

	// Store the grant ID for verification
	grantID := tokenWithGrant.Grant.ID

	// Verify the grant exists before revocation
	grant, err := client.Grant.Get(grantID)
	g.Expect(err).To(BeNil())
	g.Expect(grant).NotTo(BeNil())
	g.Expect(grant.ID).To(Equal(grantID))

	// REVOKE: Revoke the token (should delete both token and grant)
	err = client.Token.Revoke(tokenWithGrant.ID)
	g.Expect(err).To(BeNil())

	// Verify token was deleted
	_, err = client.Token.Get(tokenWithGrant.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())

	// Verify the associated grant was also deleted
	_, err = client.Grant.Get(grantID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
