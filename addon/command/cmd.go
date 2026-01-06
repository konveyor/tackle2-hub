package command

import (
	"path"

	"github.com/konveyor/tackle2-hub/addon/adapter"
	"github.com/konveyor/tackle2-hub/command"
)

var (
	addon = adapter.Addon
)

// New returns a command.
func New(p string) (cmd *command.Command) {
	cmd = &command.Command{Path: p}
	reporter := &Reporter{}
	writer := &Writer{}
	writer.reporter = reporter
	cmd.Begin = func() (err error) {
		cmd.Writer = writer
		output := path.Base(cmd.Path) + ".output"
		reporter.file, err = addon.File.Touch(output)
		if err != nil {
			return
		}
		reporter.Run(cmd.Path, cmd.Options)
		addon.Attach(reporter.file)
		return
	}
	cmd.End = func() {
		writer.End()
		if cmd.Error != nil {
			reporter.Error(cmd.Path, cmd.Error, writer.Bytes())
		} else {
			reporter.Succeeded(cmd.Path, writer.Bytes())
		}
	}
	return
}
