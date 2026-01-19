package analysis

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/analysis"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	Client     *binding.Client
	RichClient *binding.RichClient
	Analysis   analysis.Analysis
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Access REST client directly (some test API call need it)
	Client = RichClient.Client

	// Shortcut for Analysis-related RichClient methods.
	Analysis = RichClient.Analysis
}
