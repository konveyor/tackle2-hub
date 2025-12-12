package importcsv

import (
	"github.com/konveyor/tackle2-hub/shared/api"
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
					Bucket:      &api.Ref{},
					Repository: &api.Repository{
						Kind:   "git",
						URL:    "https://git-acme.local/customers.git",
						Branch: "",
						Tag:    "",
						Path:   "",
					},
					Binary: "corp.acme.demo:customers-tomcat:0.0.1-SNAPSHOT:war",
					Tags: []api.TagRef{
						{
							Name:   "Oracle",
							Source: "",
						},
						{
							Name:   "Java",
							Source: "",
						},
						{
							Name:   "RHEL 8",
							Source: "",
						},
						{
							Name:   "Tomcat",
							Source: "",
						},
					},
					BusinessService: &api.Ref{
						Name: "Retail",
					},
					Owner: &api.Ref{
						Name: "John Doe",
					},
				},
				{
					Name:        "Inventory",
					Description: "Inventory service",
					Bucket:      &api.Ref{},
					Repository: &api.Repository{
						Kind:   "git",
						URL:    "https://git-acme.local/inventory.git",
						Branch: "",
						Tag:    "",
						Path:   "",
					},
					Binary: "corp.acme.demo:inventory:0.1.1-SNAPSHOT:war",
					Tags: []api.TagRef{
						{
							Name:   "PostgreSQL",
							Source: "",
						},
						{
							Name:   "Java",
							Source: "",
						},
						{
							Name:   "RHEL 8",
							Source: "",
						},
						{
							Name:   "Quarkus",
							Source: "",
						},
					},
					BusinessService: &api.Ref{
						Name: "Retail",
					},
					Contributors: []api.Ref{
						{
							Name: "John Doe",
						},
						{
							Name: "Jane Smith",
						},
					},
				},
				{
					Name:        "Gateway",
					Description: "API Gateway",
					Bucket:      &api.Ref{},
					Repository: &api.Repository{
						Kind:   "git",
						URL:    "https://git-acme.local/gateway.git",
						Branch: "",
						Tag:    "",
						Path:   "",
					},
					Binary: "corp.acme.demo:gateway:0.1.1-SNAPSHOT:war",
					Tags: []api.TagRef{
						{
							Name:   "Java",
							Source: "",
						},
						{
							Name:   "RHEL 8",
							Source: "",
						},
						{
							Name:   "Spring Boot",
							Source: "",
						},
					},
					BusinessService: &api.Ref{
						Name: "Retail",
					},
					Owner: &api.Ref{
						Name: "John Doe",
					},
					Contributors: []api.Ref{
						{
							Name: "John Doe",
						},
						{
							Name: "Jane Smith",
						},
					},
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
