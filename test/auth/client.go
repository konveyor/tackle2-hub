package auth

import (
	"crypto/tls"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
)

// NewClient creates and returns a configured API client for testing.
func NewClient() (c *binding.RichClient) {
	c = binding.New(Settings.Addon.Hub.URL)
	c.Client.SetRetry(uint8(1))
	c.Client.Transport().TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	c.Client.Use(auth.NewBasic("admin", "admin"))
	return
}

// CreateTestClient creates an IdpClient with the specified grants and scopes for testing.
// Deletes any existing client with the same clientId first to ensure clean state.
func CreateTestClient(client *binding.RichClient, clientId string, grants []string, scopes []string) (idpClient *api.IdpClient, err error) {
	// Clean up any existing client with this ID from failed previous runs
	list, err := client.IdpClient.List()
	if err == nil {
		for _, existing := range list {
			if existing.ClientId == clientId {
				_ = client.IdpClient.Delete(existing.ID)
				break
			}
		}
	}

	// Default scopes if none provided
	if len(scopes) == 0 {
		scopes = []string{
			"openid",
			"profile",
			"email",
		}
	}

	idpClient = &api.IdpClient{
		ClientId:        clientId,
		Secret:          "test-secret",
		ApplicationType: "web",
		Grants:          grants,
		RedirectURIs: []string{
			Settings.Addon.Hub.URL + api.OIDCRoutes + "/callback",
		},
		Scopes: scopes,
	}

	err = client.IdpClient.Create(idpClient)
	return
}
