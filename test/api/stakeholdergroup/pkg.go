package stakeholdergroup

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient       *binding2.RichClient
	StakeholderGroup binding2.StakeholderGroup
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for StakeholderGroup-related RichClient methods.
	StakeholderGroup = RichClient.StakeholderGroup
}
