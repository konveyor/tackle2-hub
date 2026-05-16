package auth

import (
	"crypto/tls"

	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
	client   *binding.RichClient
)

func init() {
	client = binding.New(Settings.Addon.Hub.URL)
	client.Client.SetRetry(uint8(1))
	client.Client.Transport().TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
}
