package importcsv

import (
	"github.com/konveyor/tackle2-hub/api"
)

type TestCase struct {
	fileName             string
	ExpectedApplications []api.Application
	ExpectedDependencies []api.Dependency
}

var (
	TestCases = []TestCase{
		{
			fileName: "template_application_import.csv",
			ExpectedApplications: []api.Application{
				{
					Name:        "Customers",
					Description: "Legacy Customers management service",
				},
				{
					Name:        "Inventory",
					Description: "Inventory service",
				},
				{
					Name:        "Gateway",
					Description: "API Gateway",
				},
				{
					Name:        "Gateway",
					Description: "API Gateway",
				},
				{
					Name:        "Inventory",
					Description: "Inventory service",
				},
				{
					Name:        "Gateway",
					Description: "API Gateway",
				},
				{
					Name:        "Customers",
					Description: "Legacy Customers management service",
				},
			},
			ExpectedDependencies: []api.Dependency{
				{
					To: api.Ref{
						Name: "Inventory",
					},
					From: api.Ref{
						Name: "Gateway",
					},
				},
				{
					To: api.Ref{
						Name: "Customers",
					},
					From: api.Ref{
						Name: "Gateway",
					},
				},
			},
		},
	}
)
