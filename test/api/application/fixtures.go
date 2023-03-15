package application

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/testclient"
)

type TestCase struct {
	Test        testclient.TestCase
	Application *api.Application
}

func Create(app *api.Application) (err error) {

	return
}
