package bucket

import (
	binding2 "github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Client     *binding2.Client
	Bucket     binding2.Bucket
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Access REST client directly
	Client = RichClient.Client

	// Shortcut for Bucket related RichClient methods.
	Bucket = RichClient.Bucket
}
