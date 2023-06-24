package importcsv

import (
	"github.com/konveyor/tackle2-hub/api"
)

type TestCase struct {
	FileName             string
	ExpectedApplications []api.Application
	ExpectedDependencies []api.Dependency
}

var (
	TestCases = []TestCase{
		{
			FileName: "template_application_import.csv",
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
