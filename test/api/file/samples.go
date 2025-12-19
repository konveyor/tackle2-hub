package file

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests.
var (
	TextYAML = api.File{
		Name: "addon.yml",
		Path: "./data/addon.yml",
	}
	BinaryPNG = api.File{
		Name: "konveyor_header.png",
		Path: "./data/konveyor_header.png",
	}
	Empty = api.File{
		Name: "empty",
		Path: "./data/empty",
	}

	Samples = []api.File{TextYAML, BinaryPNG, Empty}
)
