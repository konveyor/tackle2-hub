package migrationwave

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient       *binding2.RichClient
	MigrationWave    binding2.MigrationWave
	Application      binding2.Application
	Stakeholder      binding2.Stakeholder
	StakeholderGroup binding2.StakeholderGroup
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for MigrationWave-related RichClient methods.
	MigrationWave = RichClient.MigrationWave

	// Shortcut for Application-related RichClient methods.
	Application = RichClient.Application

	// Shortcut for StakeHolder-related RichClient methods.
	Stakeholder = RichClient.Stakeholder

	// Shortcut for StakeHolderGroup-related RichClient methods.
	StakeholderGroup = RichClient.StakeholderGroup
}
