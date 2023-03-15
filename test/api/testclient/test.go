package testclient

import "github.com/konveyor/tackle2-hub/api"

type TestCase struct {
	Name        string
	Subject     interface{}
	ShouldError bool

	Application *api.Application
}
