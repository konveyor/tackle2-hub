package migrationwave

import (
	"time"

	"github.com/konveyor/tackle2-hub/shared/api"
)

var Samples = []api.MigrationWave{
	{
		Name:      "MigrationWaves",
		StartDate: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Add(30 * time.Minute),
		Applications: []api.Ref{
			{
				Name: "Sample Application",
			},
		},
		Stakeholders: []api.Ref{
			{
				Name: "Sample Stakeholders",
			},
		},
		StakeholderGroups: []api.Ref{
			{
				Name: "Sample Stakeholders Groups",
			},
		},
	},
}
