package importcsv

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

type TestCase struct {
	FileName             string
	ExpectedApplications []api2.Application
	ExpectedDependencies []api2.Dependency
}

var (
	TestCases = []TestCase{
		{
			FileName: "template_application_import.csv",
			ExpectedApplications: []api2.Application{
				{
					Name:        "Customers",
					Description: "Legacy Customers management service",
					Bucket:      &api2.Ref{},
					Repository: &api2.Repository{
						Kind:   "git",
						URL:    "https://git-acme.local/customers.git",
						Branch: "",
						Tag:    "",
						Path:   "",
					},
					Binary: "corp.acme.demo:customers-tomcat:0.0.1-SNAPSHOT:war",
					Tags: []api2.TagRef{
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
					BusinessService: &api2.Ref{
						Name: "Retail",
					},
					Owner: &api2.Ref{
						Name: "John Doe",
					},
				},
				{
					Name:        "Inventory",
					Description: "Inventory service",
					Bucket:      &api2.Ref{},
					Repository: &api2.Repository{
						Kind:   "git",
						URL:    "https://git-acme.local/inventory.git",
						Branch: "",
						Tag:    "",
						Path:   "",
					},
					Binary: "corp.acme.demo:inventory:0.1.1-SNAPSHOT:war",
					Tags: []api2.TagRef{
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
					BusinessService: &api2.Ref{
						Name: "Retail",
					},
					Contributors: []api2.Ref{
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
					Bucket:      &api2.Ref{},
					Repository: &api2.Repository{
						Kind:   "git",
						URL:    "https://git-acme.local/gateway.git",
						Branch: "",
						Tag:    "",
						Path:   "",
					},
					Binary: "corp.acme.demo:gateway:0.1.1-SNAPSHOT:war",
					Tags: []api2.TagRef{
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
					BusinessService: &api2.Ref{
						Name: "Retail",
					},
					Owner: &api2.Ref{
						Name: "John Doe",
					},
					Contributors: []api2.Ref{
						{
							Name: "John Doe",
						},
						{
							Name: "Jane Smith",
						},
					},
				},
			},
			ExpectedDependencies: []api2.Dependency{
				{
					To: api2.Ref{
						Name: "Inventory",
					},
					From: api2.Ref{
						Name: "Gateway",
					},
				},
				{
					To: api2.Ref{
						Name: "Customers",
					},
					From: api2.Ref{
						Name: "Gateway",
					},
				},
			},
		},
	}
)
