package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-jose/go-jose/v4"
	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
	. "github.com/onsi/gomega"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"golang.org/x/oauth2"
)

// mockRelyingParty implements rp.RelyingParty for testing.
type mockRelyingParty struct {
	endSessionEndpoint string
}

func (m *mockRelyingParty) OAuthConfig() *oauth2.Config                  { return &oauth2.Config{} }
func (m *mockRelyingParty) Issuer() string                               { return "" }
func (m *mockRelyingParty) IsPKCE() bool                                 { return false }
func (m *mockRelyingParty) CookieHandler() *httphelper.CookieHandler      { return nil }
func (m *mockRelyingParty) HttpClient() *http.Client                     { return nil }
func (m *mockRelyingParty) IsOAuth2Only() bool                           { return false }
func (m *mockRelyingParty) Signer() jose.Signer                         { return nil }
func (m *mockRelyingParty) GetEndSessionEndpoint() string                { return m.endSessionEndpoint }
func (m *mockRelyingParty) GetRevokeEndpoint() string                    { return "" }
func (m *mockRelyingParty) UserinfoEndpoint() string                     { return "" }
func (m *mockRelyingParty) GetDeviceAuthorizationEndpoint() string       { return "" }
func (m *mockRelyingParty) IDTokenVerifier() *rp.IDTokenVerifier         { return nil }
func (m *mockRelyingParty) ErrorHandler() func(http.ResponseWriter, *http.Request, string, string, string) {
	return nil
}
func (m *mockRelyingParty) Logger(context.Context) (*slog.Logger, bool) { return nil, false }

// TestEndSessionURL tests building the upstream IdP end_session URL.
func TestEndSessionURL(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/logout",
		},
	}

	logoutURL, ok, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(ok).To(BeTrue())
	u, pErr := url.Parse(logoutURL)
	g.Expect(pErr).To(BeNil())
	g.Expect(u.Scheme).To(Equal("https"))
	g.Expect(u.Host).To(Equal("keycloak.example.com"))
	g.Expect(u.Query().Get("client_id")).To(Equal("hub-client"))
	g.Expect(u.Query().Get("post_logout_redirect_uri")).To(Equal("https://app.example.com/"))
}

// TestEndSessionURLNoRedirect tests building the URL without a post_logout_redirect_uri.
func TestEndSessionURLNoRedirect(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/logout",
		},
	}

	logoutURL, ok, err := h.EndSessionURL("")
	g.Expect(err).To(BeNil())
	g.Expect(ok).To(BeTrue())
	u, pErr := url.Parse(logoutURL)
	g.Expect(pErr).To(BeNil())
	g.Expect(u.Query().Get("client_id")).To(Equal("hub-client"))
	g.Expect(u.Query().Get("post_logout_redirect_uri")).To(Equal(""))
}

// TestEndSessionURLExistingQuery tests that existing query parameters
// on the end_session_endpoint are preserved.
func TestEndSessionURLExistingQuery(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/logout?foo=bar",
		},
	}

	logoutURL, ok, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(ok).To(BeTrue())
	u, pErr := url.Parse(logoutURL)
	g.Expect(pErr).To(BeNil())
	g.Expect(u.Query().Get("foo")).To(Equal("bar"))
	g.Expect(u.Query().Get("client_id")).To(Equal("hub-client"))
	g.Expect(u.Query().Get("post_logout_redirect_uri")).To(Equal("https://app.example.com/"))
}

// TestEndSessionURLDisabled tests that EndSessionURL returns false when IdP is disabled.
func TestEndSessionURLDisabled(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled: false,
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "https://keycloak.example.com/logout",
		},
	}

	_, ok, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(ok).To(BeFalse())
}

// TestEndSessionURLNoEndpoint tests that EndSessionURL returns false
// when the IdP has no end_session_endpoint.
func TestEndSessionURLNoEndpoint(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{
		rpClient: &mockRelyingParty{
			endSessionEndpoint: "",
		},
	}

	_, ok, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).To(BeNil())
	g.Expect(ok).To(BeFalse())
}

// TestEndSessionURLNoClient tests that EndSessionURL returns an error
// when the RP client cannot be initialized.
func TestEndSessionURLNoClient(t *testing.T) {
	g := NewGomegaWithT(t)

	savedFederated := *federated
	defer func() { *federated = savedFederated }()

	federated.Idp = as.IdentityProvider{
		Enabled:  true,
		ClientId: "hub-client",
	}

	h := &FedIdpHandler{}

	_, ok, err := h.EndSessionURL("https://app.example.com/")
	g.Expect(err).NotTo(BeNil())
	g.Expect(ok).To(BeFalse())
}
