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

// TestBearerHeader tests Bearer Header method.
func TestBearerHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create Bearer with mock tokens (skip RP client initialization)
	bearer := &Bearer{
		accessToken:  "test-access-token",
		refreshToken: "test-refresh-token",
	}

	header := bearer.Header()
	g.Expect(header).To(Equal("Bearer test-access-token"))
}

// TestBearerLoginNoRefreshToken tests Login error when not authenticated.
func TestBearerLoginNoRefreshToken(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create Bearer with no tokens
	bearer := &Bearer{}

	err := bearer.Login()
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("not authenticated"))
}
