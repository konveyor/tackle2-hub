package binding

import (
	"crypto/tls"
	"os"

	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
	client2 "github.com/konveyor/tackle2-hub/shared/binding/client"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

const (
	User     = "BINDING_USER"
	Password = "BINDING_PASSWORD"
)

var (
	Settings   = &settings.Settings
	client     *binding.RichClient
	authMethod client2.AuthMethod
)

func init() {
	client = binding.New(Settings.Addon.Hub.URL)
	client.Client.SetRetry(uint8(1))
	client.Client.Transport().TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	user := os.Getenv(User)
	password := os.Getenv(Password)
	if user == "" {
		user = "admin"
	}
	if password == "" {
		password = "admin"
	}
	authMethod = auth.NewBasic(user, password)
	client.Client.Use(authMethod)
}
