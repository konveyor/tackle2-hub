package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestServiceAccount(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get available scopes from the hub
	scopes, err := client.Scope.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(scopes)).Should(BeNumerically(">=", 2))

	// Create roles for the service account to reference
	role1 := &api.Role{
		Name: "sa-role-1",
		Scopes: []string{
			scopes[0].Name,
			scopes[1].Name,
		},
	}
	err = client.Role.Create(role1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Role.Delete(role1.ID)
	})

	role2 := &api.Role{
		Name: "sa-role-2",
		Scopes: []string{
			scopes[0].Name,
		},
	}
	err = client.Role.Create(role2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Role.Delete(role2.ID)
	})

	// Get seeded service accounts
	seeded, err := client.ServiceAccount.List()
	g.Expect(err).To(BeNil())

	// Define the service account to create
	sa := &api.ServiceAccount{
		Name: "test-sa",
		Roles: []api.Ref{
			{ID: role1.ID},
			{ID: role2.ID},
		},
	}

	// CREATE: Create the service account
	err = client.ServiceAccount.Create(sa)
	g.Expect(err).To(BeNil())
	g.Expect(sa.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.ServiceAccount.Delete(sa.ID)
	})

	// GET: List service accounts
	list, err := client.ServiceAccount.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))

	// GET: Retrieve the service account and verify it matches
	retrieved, err := client.ServiceAccount.Get(sa.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.Subject).NotTo(BeEmpty())
	g.Expect(retrieved.Name).To(Equal(sa.Name))

	// Verify roles are associated
	g.Expect(len(retrieved.Roles)).To(Equal(2))
	g.Expect(retrieved.Roles).To(ContainElement(api.Ref{ID: role1.ID, Name: role1.Name}))
	g.Expect(retrieved.Roles).To(ContainElement(api.Ref{ID: role2.ID, Name: role2.Name}))

	// UPDATE: Modify the service account
	sa.Name = "updated-sa"
	sa.Roles = []api.Ref{
		{ID: role1.ID},
	}

	err = client.ServiceAccount.Update(sa)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.ServiceAccount.Get(sa.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(len(updated.Roles)).To(Equal(1))
	g.Expect(updated.Roles).To(ContainElement(api.Ref{ID: role1.ID, Name: role1.Name}))

	// DELETE: Remove the service account
	err = client.ServiceAccount.Delete(sa.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.ServiceAccount.Get(sa.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

func TestServiceAccountToken(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get available scopes from the hub
	scopes, err := client.Scope.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(scopes)).Should(BeNumerically(">=", 1))

	// Create a role for the service account
	role := &api.Role{
		Name: "sa-token-role",
		Scopes: []string{
			scopes[0].Name,
		},
	}
	err = client.Role.Create(role)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Role.Delete(role.ID)
	})

	// Create a service account
	sa := &api.ServiceAccount{
		Name: "token-test-sa",
		Roles: []api.Ref{
			{ID: role.ID},
		},
	}
	err = client.ServiceAccount.Create(sa)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.ServiceAccount.Delete(sa.ID)
	})

	// CREATE: Create a token for the service account
	selected := client.ServiceAccount.Select(sa.ID)
	pat := &api.PAT{}
	err = selected.Token.Create(pat)
	g.Expect(err).To(BeNil())
	g.Expect(pat.ID).NotTo(BeZero())
	g.Expect(pat.Token).NotTo(BeEmpty())
	t.Cleanup(func() {
		_ = client.Token.Delete(pat.ID)
	})

	// GET: Retrieve the token and verify it is linked to the SA
	token, err := client.Token.Get(pat.ID)
	g.Expect(err).To(BeNil())
	g.Expect(token).NotTo(BeNil())
	g.Expect(token.Kind).To(Equal(api.TokenKindAPIKey))
	g.Expect(token.ServiceAccount).NotTo(BeNil())
	g.Expect(token.ServiceAccount.ID).To(Equal(sa.ID))

	// GET: Retrieve the SA and verify token is listed
	retrieved, err := client.ServiceAccount.Get(sa.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(retrieved.Tokens)).To(Equal(1))
	g.Expect(retrieved.Tokens[0].ID).To(Equal(pat.ID))

	// DELETE: Revoke the token
	err = client.Token.Revoke(pat.ID)
	g.Expect(err).To(BeNil())

	// Verify token was revoked
	_, err = client.Token.Get(pat.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
