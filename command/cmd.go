/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/jortel/go-utils/logr"
)

var Log = logr.WithName("command")

// New returns a command.
func New(path string) (cmd *Command) {
	cmd = &Command{Path: path}
	return
}

// Command execution.
type Command struct {
	Options Options
	Path    string
	Dir     string
	Env     []string
	Writer  io.Writer
	Error   error
}

// Run executes the command.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) Run() (err error) {
	err = r.RunWith(context.TODO())
	return
}

// RunWith executes the command with context.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) RunWith(ctx context.Context) (err error) {
	defer func() {
		r.Error = err
		Log.Info(r.String())
	}()
	if r.Writer == nil {
		r.Writer = &bytes.Buffer{}
	}
	cmd := exec.CommandContext(ctx, r.Path, r.Options...)
	cmd.Dir = r.Dir
	cmd.Env = r.Env
	cmd.Stdout = r.Writer
	cmd.Stderr = r.Writer
	err = cmd.Start()
	if err != nil {
		return
	}
	err = cmd.Wait()
	return
}

// Output returns the command output.
func (r *Command) Output() (b []byte) {
	if reader, cast := r.Writer.(io.Reader); cast {
		b, _ = io.ReadAll(reader)
	}
	return
}

func (r *Command) String() (s string) {
	parts := []string{
		"[CMD] ",
		r.Path,
		strings.Join(r.Options, " "),
	}
	if r.Error != nil {
		parts = append(parts, "\nFAILED:", r.Error.Error())
	}
	output := r.Output()
	if len(output) > 0 {
		parts = append(parts, "\noutput:\n", string(output))
	}
	s = strings.Join(parts, " ")
	return
}
