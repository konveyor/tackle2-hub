/*
Package ssh provides a SSH related functionality.
*/
package ssh

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/command"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/shared/ssh"
)

var (
	Settings = &settings.Settings
)

func init() {
	ssh.NewCommand = command.New
	ssh.Log = logr.New("ssh", Settings.Log.SSH)
}

type Agent = ssh.Agent
type Key = ssh.Key
