package questionnaire

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient    *binding.RichClient
	Questionnaire binding.Questionnaire
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Questionnaire-related RichClient methods.
	Questionnaire = RichClient.Questionnaire
}
