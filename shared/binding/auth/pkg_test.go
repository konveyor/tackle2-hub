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

// TestBearerTokenMethods tests Bearer token storage and retrieval.
func TestBearerTokenMethods(t *testing.T) {
	g := NewGomegaWithT(t)

	bearer := NewBearer("http://localhost:8080", "test-client")

	// Use method should set access token
	bearer.Use("test-access-token")
	g.Expect(bearer.Token()).To(Equal("test-access-token"))

	// Header should include Bearer prefix
	header := bearer.Header()
	g.Expect(header).To(Equal("Bearer test-access-token"))
}

// TestBearerSetTransport tests Bearer transport configuration.
func TestBearerSetTransport(t *testing.T) {
	g := NewGomegaWithT(t)

	bearer := NewBearer("http://localhost:8080", "test-client")

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
