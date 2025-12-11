package migrationwave

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient       *binding.RichClient
	MigrationWave    binding.MigrationWave
	Application      binding.Application
	Stakeholder      binding.Stakeholder
	StakeholderGroup binding.StakeholderGroup
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
