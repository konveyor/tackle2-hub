package manifest

import (
	"github.com/konveyor/tackle2-hub/api"
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
		},
		Secret: api.Map{
			"key":      "ABCDEF",
			"user":     "Elmer",
			"password": "1234",
		},
	}

	Samples = []api.Manifest{Base}
)
