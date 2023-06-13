package importCSV

import (
	"github.com/konveyor/tackle2-hub/api"
)

type TestCase struct {
	fileName             string
	ExpectedApplications []api.Application
	ExpectedDependencies []api.Dependency
}

var (
	Applications = []api.Application{
		{
			Name:        "Gateway",
			Description: "Gateway application",
		},
		{
			Name:        "Inventory",
			Description: "Inventory application",
		},
	}
	Dependencies = []api.Dependency{
		{
			To: api.Ref{
				Name: "Gateway",
			},
			From: api.Ref{
				Name: "Inventory",
			},
		},
		{
			To: api.Ref{
				Name: "Inventory",
			},
			From: api.Ref{
				Name: "Customers",
			},
		},
	}
	Samples = []TestCase{
		{
			fileName:             "template_application_import.csv",
			ExpectedApplications: Applications,
			ExpectedDependencies: Dependencies,
		},
	}
)
