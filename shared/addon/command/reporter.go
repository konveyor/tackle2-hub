package command

import (
	"strings"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/command"
)

// Verbosity.
const (
	// Disabled reports: NOTHING.
	Disabled = -2
	// Error reports: error.
	Error = -1
	// Default reports: error, started, succeeded.
	Default = 0
	// LiveOutput reports: error, started, succeeded, output (live).
	LiveOutput = 1
)

// Reporter provides integration with the task Report.Activity.
type Reporter struct {
	Verbosity int
	file      *api.File
	index     int
}

// Run reports command started in task Report.Activity.
func (r *Reporter) Run(path string, options command.Options) {
	switch r.Verbosity {
	case Disabled:
	case Error:
	case Default,
		LiveOutput:
		addon.Activity(
			"[CMD] Running: %s %s",
			path,
			strings.Join(options, " "))
	}
}

// Succeeded reports command succeeded in task Report.Activity.
func (r *Reporter) Succeeded(path string, output []byte) {
	switch r.Verbosity {
	case Disabled:
	case Error:
	case Default:
		addon.Activity(
			"[CMD] %s succeeded.",
			path)
		r.append(output)
	case LiveOutput:
		addon.Activity(
			"[CMD] %s succeeded.",
			path)
	}
}

// Error reports command failed in task Report.Activity.
func (r *Reporter) Error(path string, err error, output []byte) {
	if len(output) == 0 {
		return
	}
	switch r.Verbosity {
	case Disabled:
	case Error,
		Default:
		addon.Activity(
			"[CMD] %s failed: %s",
			path,
			err.Error())
		r.append(output)
	case LiveOutput:
		addon.Activity(
			"[CMD] %s failed: %s.",
			path,
			err.Error())
	}
}

// Output reports command output.
func (r *Reporter) Output(buffer []byte) (reported int) {
	switch r.Verbosity {
	case Disabled:
	case Error:
	case Default:
	case LiveOutput:
		if r.index >= len(buffer) {
			return
		}
		batch := buffer[r.index:]
		reported = len(batch)
		if reported > 0 {
			r.index += reported
			r.append(batch)
		}
	}
	return
}

// append output.
func (r *Reporter) append(batch []byte) {
	if r.file == nil {
		return
	}
	err := addon.File.Patch(r.file.ID, batch)
	if err != nil {
		panic(err)
	}
}
