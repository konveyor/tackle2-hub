package settings

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/settings"
	. "github.com/onsi/gomega"
)

var Settings = &settings.Settings

// TestFederatedLoadClients tests loading IdpClients from CRDs in disconnected mode.
func TestFederatedLoadClients(t *testing.T) {
	g := NewGomegaWithT(t)

	// Save and restore settings
	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
	}()
	Settings.Disconnected = true

	federated := &Federated{}
	err := federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	g.Expect(federated.Clients).To(HaveLen(4))

	// Verify web-ui client
	var webUI *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "web-ui" {
			webUI = &federated.Clients[i]
			break
		}
	}
	g.Expect(webUI).NotTo(BeNil())
	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(webUI.ApplicationType).To(Equal("web"))

	// Verify kantra client
	var kantra *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "kantra" {
			kantra = &federated.Clients[i]
			break
		}
	}
	g.Expect(kantra).NotTo(BeNil())
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kantra.ApplicationType).To(Equal("native"))

	// Verify kai-ide client
	var kaiIDE *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "kai-ide" {
			kaiIDE = &federated.Clients[i]
			break
		}
	}
	g.Expect(kaiIDE).NotTo(BeNil())
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))
	g.Expect(kaiIDE.ApplicationType).To(Equal("native"))

	// Verify web-ui-with-secret client
	var confidential *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "web-ui-with-secret" {
			confidential = &federated.Clients[i]
			break
		}
	}
	g.Expect(confidential).NotTo(BeNil())
	g.Expect(confidential.ID).To(Equal(uint(4)))
	g.Expect(confidential.ApplicationType).To(Equal("web"))
}

// TestFederatedGetClientsWithSecret tests secret resolution for clients with clientSecret reference.
func TestFederatedGetClientsWithSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
	}()
	Settings.Disconnected = true

	federated := &Federated{}
	err := federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	// Find web-ui-with-secret client (confidential client with secret reference)
	var confidential *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "web-ui-with-secret" {
			confidential = &federated.Clients[i]
			break
		}
	}
	g.Expect(confidential).NotTo(BeNil())
	g.Expect(confidential.Secret).To(Equal("test-secret-value"))
}

// TestFederatedGetClientsPublic tests public clients work without secrets.
func TestFederatedGetClientsPublic(t *testing.T) {
	g := NewGomegaWithT(t)

	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
	}()
	Settings.Disconnected = true

	federated := &Federated{}
	err := federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	// Find kantra client (public)
	var kantra *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "kantra" {
			kantra = &federated.Clients[i]
			break
		}
	}
	g.Expect(kantra).NotTo(BeNil())
	g.Expect(kantra.Secret).To(BeEmpty())

	// Find kai-ide client (public)
	var kaiIDE *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "kai-ide" {
			kaiIDE = &federated.Clients[i]
			break
		}
	}
	g.Expect(kaiIDE).NotTo(BeNil())
	g.Expect(kaiIDE.Secret).To(BeEmpty())
}

// TestFederatedGetClientsFields tests all client fields are loaded correctly.
func TestFederatedGetClientsFields(t *testing.T) {
	g := NewGomegaWithT(t)

	originalDisconnected := Settings.Disconnected
	defer func() {
		Settings.Disconnected = originalDisconnected
	}()
	Settings.Disconnected = true

	federated := &Federated{}
	err := federated.Load("konveyor-tackle")
	g.Expect(err).To(BeNil())

	// Find web-ui client
	var webUI *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "web-ui" {
			webUI = &federated.Clients[i]
			break
		}
	}
	g.Expect(webUI).NotTo(BeNil())
	g.Expect(webUI.ID).To(Equal(uint(1)))
	g.Expect(webUI.ClientId).To(Equal("web-ui"))
	g.Expect(webUI.ApplicationType).To(Equal("web"))
	g.Expect(webUI.Grants).To(ContainElement("authorization_code"))
	g.Expect(webUI.Grants).To(ContainElement("refresh_token"))
	g.Expect(webUI.Scopes).To(ContainElement("openid"))
	g.Expect(webUI.Scopes).To(ContainElement("profile"))

	// Find kantra client
	var kantra *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "kantra" {
			kantra = &federated.Clients[i]
			break
		}
	}
	g.Expect(kantra).NotTo(BeNil())
	g.Expect(kantra.ID).To(Equal(uint(2)))
	g.Expect(kantra.ApplicationType).To(Equal("native"))
	g.Expect(kantra.Grants).To(ContainElement("urn:ietf:params:oauth:grant-type:device_code"))
	g.Expect(kantra.Scopes).To(ContainElement("openid"))

	// Find kai-ide client
	var kaiIDE *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "kai-ide" {
			kaiIDE = &federated.Clients[i]
			break
		}
	}
	g.Expect(kaiIDE).NotTo(BeNil())
	g.Expect(kaiIDE.ID).To(Equal(uint(3)))
	g.Expect(kaiIDE.ApplicationType).To(Equal("native"))
	g.Expect(kaiIDE.Grants).To(ContainElement("authorization_code"))
	g.Expect(kaiIDE.Scopes).To(ContainElement("openid"))
}

// TestIdentityProviderInject tests template variable injection.
func TestIdentityProviderInject(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "https://${issuer.host}/auth",
		RedirectURI: "${issuer.proto}://${issuer.host}:${issuer.port}/callback",
	}

	issuer := "https://hub.example.com:8443/oidc"
	idp.Inject(issuer)

	g.Expect(idp.Issuer).To(Equal("https://hub.example.com/auth"))
	g.Expect(idp.RedirectURI).To(Equal("https://hub.example.com:8443/callback"))
}

// TestIdentityProviderInjectIssuerVariable tests ${issuer} template variable.
func TestIdentityProviderInjectIssuerVariable(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "${issuer}/external",
		RedirectURI: "${issuer}/callback",
	}

	issuer := "https://auth.example.com"
	idp.Inject(issuer)

	g.Expect(idp.Issuer).To(Equal("https://auth.example.com/external"))
	g.Expect(idp.RedirectURI).To(Equal("https://auth.example.com/callback"))
}

// TestIdentityProviderInjectIdempotent tests that Inject() is idempotent.
func TestIdentityProviderInjectIdempotent(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "https://${issuer.host}/auth",
		RedirectURI: "${issuer}/callback",
	}

	issuer := "https://hub.example.com/oidc"

	// First injection
	idp.Inject(issuer)
	firstIssuer := idp.Issuer
	firstRedirect := idp.RedirectURI

	// Second injection with different issuer
	idp.Inject("https://different.example.com/oidc")

	// Values should not change (idempotent)
	g.Expect(idp.Issuer).To(Equal(firstIssuer))
	g.Expect(idp.RedirectURI).To(Equal(firstRedirect))
}

// TestIdentityProviderInjectNoTemplates tests Inject() with no template variables.
func TestIdentityProviderInjectNoTemplates(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "https://static.example.com/auth",
		RedirectURI: "https://app.example.com/callback",
	}

	issuer := "https://hub.example.com/oidc"
	idp.Inject(issuer)

	// Values should remain unchanged
	g.Expect(idp.Issuer).To(Equal("https://static.example.com/auth"))
	g.Expect(idp.RedirectURI).To(Equal("https://app.example.com/callback"))
}

// TestIdentityProviderInjectAllVariables tests all template variables.
func TestIdentityProviderInjectAllVariables(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "${issuer.proto}://${issuer.host}:${issuer.port}${issuer.path}/auth",
		RedirectURI: "${issuer}/callback",
	}

	issuer := "https://hub.example.com:9443/oidc"
	idp.Inject(issuer)

	g.Expect(idp.Issuer).To(Equal("https://hub.example.com:9443/oidc/auth"))
	g.Expect(idp.RedirectURI).To(Equal("https://hub.example.com:9443/oidc/callback"))
}

// TestIdentityProviderInjectDefaultPort tests template variables with default port.
func TestIdentityProviderInjectDefaultPort(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "https://${issuer.host}:${issuer.port}/auth",
		RedirectURI: "http://${issuer.host}:${issuer.port}/callback",
	}

	// HTTPS default port (443) - url.Parse returns empty string for default ports
	issuer := "https://hub.example.com/oidc"
	idp.Inject(issuer)

	// When port is default (443 for https), ${issuer.port} expands to empty string
	g.Expect(idp.Issuer).To(Equal("https://hub.example.com:/auth"))
	g.Expect(idp.RedirectURI).To(Equal("http://hub.example.com:/callback"))
}

// TestIdentityProviderInjectConcurrent tests thread-safety of Inject().
func TestIdentityProviderInjectConcurrent(t *testing.T) {
	g := NewGomegaWithT(t)

	idp := &IdentityProvider{
		Issuer:      "https://${issuer.host}/auth",
		RedirectURI: "${issuer}/callback",
	}

	issuer := "https://hub.example.com/oidc"

	// Call Inject concurrently from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			idp.Inject(issuer)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify injection happened correctly (only once due to idempotency)
	g.Expect(idp.Issuer).To(Equal("https://hub.example.com/auth"))
	g.Expect(idp.RedirectURI).To(Equal("https://hub.example.com/oidc/callback"))
}
