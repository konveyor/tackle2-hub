package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestGrant(t *testing.T) {
	g := NewGomegaWithT(t)

	// GET: List grants
	list, err := client.Grant.List()
	g.Expect(err).To(BeNil())

	// Note: Grants are only created by OIDC authorization code flow,
	// not by API key authentication. The test suite uses API keys,
	// so there may be no grants.
	if len(list) == 0 {
		t.Skip("No grants available - test suite uses API key authentication")
		return
	}

	// GET: Retrieve the first grant by ID
	firstGrant := list[0]
	retrieved, err := client.Grant.Get(firstGrant.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(firstGrant.ID))
	g.Expect(retrieved.GrantId).To(Equal(firstGrant.GrantId))
	g.Expect(retrieved.ClientId).To(Equal(firstGrant.ClientId))
	g.Expect(retrieved.Subject).To(Equal(firstGrant.Subject))
	g.Expect(retrieved.Type).To(Equal(firstGrant.Type))

	// DELETE: Remove the grant
	err = client.Grant.Delete(firstGrant.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Grant.Get(firstGrant.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
