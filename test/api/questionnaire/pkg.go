package questionnaire

import (
	binding2 "github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient    *binding2.RichClient
	Questionnaire binding2.Questionnaire
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Questionnaire-related RichClient methods.
	Questionnaire = RichClient.Questionnaire
}
