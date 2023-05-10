package analysis

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/assert"
)

var (
	// Setup Hub API client
	Client     *binding.Client
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
	Name          string
	Application   api.Application
	Task          api.Task
	TaskData      string
	ReportContent map[string][]string
}

func getReportText(t *testing.T, tc *TC, path string) (text string) {
	// Get report file.
	dirName, err := os.MkdirTemp("/tmp", tc.Name)
	assert.Must(t, err)
	fileName := filepath.Join(dirName, filepath.Base(path))
	err = RichClient.Application.Bucket(tc.Application.ID).Get(path, dirName)
	assert.Must(t, err)
	content, err := os.ReadFile(fileName)
	assert.Must(t, err)

	// Prepare content - strip tags etc.
	tags := regexp.MustCompile(`<.*?>`)
	spaces := regexp.MustCompile(`(\t|  +|\n\t+\n)`)
	text = tags.ReplaceAllString(string(content), "")
	text = spaces.ReplaceAllString(text, "")
	return
}
