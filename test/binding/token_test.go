package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestToken(t *testing.T) {
	g := NewGomegaWithT(t)

	req := &api.TokenRequest{}
	err := client.Token.Create(req)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Token.Delete(req.ID)
	})

	// GET: List tokens
	list, err := client.Token.List()
	g.Expect(err).To(BeNil())

	// Verify we have at least one token (from the test client login)
	g.Expect(len(list)).To(BeNumerically(">", 0))

	// GET: Retrieve the first token by ID
	firstToken := list[0]
	retrieved, err := client.Token.Get(firstToken.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(firstToken.ID))
	g.Expect(retrieved.AuthId).To(Equal(firstToken.AuthId))
	g.Expect(retrieved.Kind).To(Equal(firstToken.Kind))
	g.Expect(retrieved.Subject).To(Equal(firstToken.Subject))

	// DELETE: Remove the token
	err = client.Token.Delete(firstToken.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Token.Get(firstToken.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
