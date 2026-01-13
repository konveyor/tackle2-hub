package migrationwave

import (
	"time"

	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

var Samples = []api2.MigrationWave{
	{
		Name:      "MigrationWaves",
		StartDate: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Add(30 * time.Minute),
		Applications: []api2.Ref{
			{
				Name: "Sample Application",
			},
		},
		Stakeholders: []api2.Ref{
			{
				Name: "Sample Stakeholders",
			},
		},
		StakeholderGroups: []api2.Ref{
			{
				Name: "Sample Stakeholders Groups",
			},
		},
	},
}
