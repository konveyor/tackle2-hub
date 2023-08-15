package tracker

import (
	"time"

	"github.com/konveyor/tackle2-hub/api"
)

var Samples = []api.Tracker{
	{
		Name:        "Sample tracker",
		URL:         "https://konveyor.io/test/api/tracker",
		Kind:        "jira-onprem",
		Message:     "Description of tracker",
		Connected:   true,
		LastUpdated: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		Identity: api.Ref{
			ID:   1,
			Name: "Sample Tracker Identity",
		},
		Insecure: false,
	},
	{
		Name:        "Sample tracker1",
		URL:         "https://konveyor.io/test/api/tracker1",
		Kind:        "jira-cloud",
		Message:     "Description of tracker1",
		Connected:   true,
		LastUpdated: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		Identity: api.Ref{
			ID:   2,
			Name: "Sample Tracker Identity1",
		},
		Insecure: false,
	},
}
