package manifest

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Base = api.Manifest{
		Content: api.Map{
			"name": "Test",
			"key":  "$(key)",
			"database": api.Map{
				"url":      "db.com",
				"user":     "$(user)",
				"password": "$(password)",
			},
			"description": "Connect using $(user) and $(password)",
		},
		Secret: api.Map{
			"key":      "ABCDEF",
			"user":     "Elmer",
			"password": "1234",
		},
	}
	InjectedContent = api.Map{
		"name": "Test",
		"key":  "ABCDEF",
		"database": api.Map{
			"url":      "db.com",
			"user":     "Elmer",
			"password": "1234",
		},
		"description": "Connect using Elmer and 1234",
	}
)
