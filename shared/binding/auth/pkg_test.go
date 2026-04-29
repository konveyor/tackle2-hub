package auth

import (
	"testing"

	. "github.com/onsi/gomega"
)

// TestBasic tests Basic authenticator.
func TestBasic(t *testing.T) {
	g := NewGomegaWithT(t)

	auth := NewBasic("testuser", "testpass")

	// Login should be no-op
	err := auth.Login()
	g.Expect(err).To(BeNil())

	// Header should return basic auth header
	header := auth.Header()
	g.Expect(header).To(Equal("Basic dGVzdHVzZXI6dGVzdHBhc3M="))
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
