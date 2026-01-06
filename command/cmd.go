/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"context"
	"io"
	"os/exec"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
)

var (
	New func(string) *Command
	Log = logr.New("command", 0)
)

func init() {
	New = func(path string) (cmd *Command) {
		cmd = &Command{Path: path}
		return
	}
}

// Command execution.
type Command struct {
	Options Options
	Path    string
	Dir     string
	Env     []string
	Writer  io.Writer
	Error   error
	Begin   func() error
	End     func()
}

// Run executes the command.
func (r *Command) Run() (err error) {
	err = r.RunWith(context.TODO())
	return
}

// RunWith executes the command with context.
func (r *Command) RunWith(ctx context.Context) (err error) {
	defer func() {
		r.Error = err
		Log.V(1).Info(r.String())
		if r.End != nil {
			r.End()
		}
	}()
	if r.Begin != nil {
		err = r.Begin()
		if err != nil {
			return
		}
	}
	if r.Writer == nil {
		r.Writer = &Buffer{}
	}
	cmd := exec.CommandContext(ctx, r.Path, r.Options...)
	cmd.Dir = r.Dir
	cmd.Env = r.Env
	cmd.Stdout = r.Writer
	cmd.Stderr = r.Writer
	err = cmd.Start()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Output returns the command output.
func (r *Command) Output() (b []byte) {
	var err error
	if seeker, cast := r.Writer.(io.Seeker); cast {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			Log.Error(err, "")
		}
	}
	if reader, cast := r.Writer.(io.Reader); cast {
		b, err = io.ReadAll(reader)
		if err != nil {
			Log.Error(err, "")
		}
	}
	return
}

// String returns a string representation.
func (r *Command) String() (s string) {
	parts := []string{
		"[CMD] ",
		r.Path,
		strings.Join(r.Options, " "),
	}
	if r.Error == nil {
		parts = append(parts, "\nSUCCEEDED")
	} else {
		parts = append(parts, "\nFAILED:", r.Error.Error())
	}
	if Log.V(2).Enabled() {
		output := r.Output()
		if len(output) > 0 {
			parts = append(parts, "\noutput:\n", string(output))
		}
	}
	s = strings.Join(parts, " ")
	return
}
