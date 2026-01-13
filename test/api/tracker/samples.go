package tracker

import (
	"time"

	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

var Samples = []api2.Tracker{
	{
		Name:        "Sample tracker",
		URL:         "https://konveyor.io/test/api/tracker",
		Kind:        "jira-onprem",
		Message:     "Description of tracker",
		LastUpdated: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		Identity: api2.Ref{
			Name: "Sample Tracker Identity",
		},
		Insecure: false,
	},
	{
		Name:        "Sample tracker1",
		URL:         "https://konveyor.io/test/api/tracker1",
		Kind:        "jira-cloud",
		Message:     "Description of tracker1",
		LastUpdated: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		Identity: api2.Ref{
			Name: "Sample Tracker Identity1",
		},
		Insecure: false,
	},
}
