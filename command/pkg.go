/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/shared/command"
)

var (
	Settings = &settings.Settings
)

type Command = command.Command
type Options = command.Options

func init() {
	command.Log = logr.New("command", Settings.Log.Command)
}

// New returns a command.
func New(path string) (cmd *Command) {
	cmd = command.New(path)
	return
}
