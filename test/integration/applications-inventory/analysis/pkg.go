package analysis

import (
	"time"

	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = c.Client

	// Analysis waiting loop 5 minutes (60 * 5s)
	Retry = 60
	Wait  = 5 * time.Second
)

// Test cases for Application Analysis.
type TC struct {
	Name        string
	Application api.Application
	Task        api.Task
	TaskData    string
}
