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

	g.Expect(federated.Clients).To(HaveLen(3))

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

	// Find web-ui client
	var webUI *IdpClient
	for i := range federated.Clients {
		if federated.Clients[i].ClientId == "web-ui" {
			webUI = &federated.Clients[i]
			break
		}
	}
	g.Expect(webUI).NotTo(BeNil())

	// Verify secret was resolved from Kubernetes Secret
	g.Expect(webUI.Secret).To(Equal("test-secret-value"))
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
