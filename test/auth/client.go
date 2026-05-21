package auth

import (
	"crypto/tls"

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
