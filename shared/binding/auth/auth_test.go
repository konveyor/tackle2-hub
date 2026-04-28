package auth

import (
	"testing"

	. "github.com/onsi/gomega"
)

// TestAPIKey tests APIKey authenticator.
func TestAPIKey(t *testing.T) {
	g := NewGomegaWithT(t)

	auth := NewAPIKey("test-api-key-123")

	// Login should be no-op
	err := auth.Login()
	g.Expect(err).To(BeNil())

	// Header should return bearer token
	header := auth.Header()
	g.Expect(header).To(Equal("Bearer test-api-key-123"))
}

// TestOIDCHeader tests OIDC Header method.
func TestOIDCHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create OIDC with mock tokens (skip RP client initialization)
	oidc := &OIDC{
		accessToken:  "test-access-token",
		refreshToken: "test-refresh-token",
	}

	header := oidc.Header()
	g.Expect(header).To(Equal("Bearer test-access-token"))
}

// TestOIDCLoginNoRefreshToken tests Login error when not authenticated.
func TestOIDCLoginNoRefreshToken(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create OIDC with no tokens
	oidc := &OIDC{}

	err := oidc.Login()
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not authenticated"))
}
