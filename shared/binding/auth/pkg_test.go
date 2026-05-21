package auth

import (
	"net/http"
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

// TestOIDCHeader tests OIDC Header method.
func TestOIDCHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create OIDC with mock tokens (skip RP client initialization)
	bearer := &OIDC{
		accessToken:  "test-access-token",
		refreshToken: "test-refresh-token",
	}

	header := bearer.Header()
	g.Expect(header).To(Equal("OIDC test-access-token"))
}

// TestOIDCTokenMethods tests OIDC token storage and retrieval.
func TestOIDCTokenMethods(t *testing.T) {
	g := NewGomegaWithT(t)

	bearer := NewOIDC("http://localhost:8080", "test-client")

	// Use method should set access token
	bearer.Use("test-access-token")
	g.Expect(bearer.Token()).To(Equal("test-access-token"))

	// Header should include OIDC prefix
	header := bearer.Header()
	g.Expect(header).To(Equal("OIDC test-access-token"))
}

// TestOIDCSetTransport tests OIDC transport configuration.
func TestOIDCSetTransport(t *testing.T) {
	g := NewGomegaWithT(t)

	bearer := NewOIDC("http://localhost:8080", "test-client")

	// SetTransport should update httpClient transport
	transport := &http.Transport{
		MaxIdleConns: 100,
	}
	bearer.SetTransport(transport)
	g.Expect(bearer.httpClient.Transport).To(Equal(transport))
}

// TestBasicSetTransport tests Basic transport configuration.
func TestBasicSetTransport(t *testing.T) {
	g := NewGomegaWithT(t)

	basic := NewBasic("testuser", "testpass")

	// SetTransport should be a no-op
	transport := &http.Transport{}
	basic.SetTransport(transport)
	// No error means success (no-op method)
	g.Expect(true).To(BeTrue())
}

// TestNoAuthSetTransport tests NoAuth transport configuration.
func TestNoAuthSetTransport(t *testing.T) {
	g := NewGomegaWithT(t)

	noauth := &NoAuth{}

	// SetTransport should be a no-op
	transport := &http.Transport{}
	noauth.SetTransport(transport)
	// No error means success (no-op method)
	g.Expect(true).To(BeTrue())
}
