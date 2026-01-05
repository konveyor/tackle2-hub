package schema

import (
	"github.com/konveyor/tackle2-hub/internal/settings"
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
	"github.com/konveyor/tackle2-hub/shared/binding"
)

var (
	RichClient *binding.RichClient
	Settings   = &settings.Settings
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()
	err := Settings.Load()
	if err != nil {
		panic(err)
	}
}
