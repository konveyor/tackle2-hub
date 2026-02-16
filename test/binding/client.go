package binding

import (
	"os"

	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

const (
	User     = "USER"
	Password = "PASSWORD"
)

var (
	Settings = &settings.Settings
	client   *binding.RichClient
)

func init() {
	client = binding.New(Settings.Addon.Hub.URL)
	client.Client.SetRetry(uint8(1))
	user := os.Getenv(User)
	password := os.Getenv(Password)
	if user == "" || password == "" {
		return
	}
	err := client.Login(user, password)
	if err != nil {
		panic(err)
	}
}
