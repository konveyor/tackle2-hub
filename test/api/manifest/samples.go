package manifest

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Base = api2.Manifest{
		Content: api2.Map{
			"name": "Test",
			"key":  "$(key)",
			"database": api2.Map{
				"url":      "db.com",
				"user":     "$(user)",
				"password": "$(password)",
			},
			"description": "Connect using $(user) and $(password)",
		},
		Secret: api2.Map{
			"key":      "ABCDEF",
			"user":     "Elmer",
			"password": "1234",
		},
	}
	InjectedContent = api2.Map{
		"name": "Test",
		"key":  "ABCDEF",
		"database": api2.Map{
			"url":      "db.com",
			"user":     "Elmer",
			"password": "1234",
		},
		"description": "Connect using Elmer and 1234",
	}
)
