package profile

import (
	binding2 "github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient      *binding2.RichClient
	AnalysisProfile binding2.AnalysisProfile
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Profile-related RichClient methods.
	AnalysisProfile = RichClient.AnalysisProfile
}
