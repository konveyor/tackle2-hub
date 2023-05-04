package analysis

import (
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client *binding.Client
	RichClient *binding.RichClient

	// Analysis waiting loop 5 minutes (60 * 5s)
	Retry = 60
	Wait  = 5 * time.Second
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Access REST client directly (some test API call need it)
	Client = RichClient.Client()
}

// Test cases for Application Analysis.
type TC struct {
	Name        string
	Application api.Application
	Task        api.Task
	TaskData    string
}
