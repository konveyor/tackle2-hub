package importcsv

import (
	binding2 "github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding2.RichClient
	Client      *binding2.Client
	Application binding2.Application
	Dependency  binding2.Dependency
	Stakeholder binding2.Stakeholder
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Access REST client directly
	Client = RichClient.Client

	// Access Application directly
	Application = RichClient.Application

	// Access Dependency directly
	Dependency = RichClient.Dependency

	// Access Stakeholder directly
	Stakeholder = RichClient.Stakeholder
}
