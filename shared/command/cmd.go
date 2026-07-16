/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"context"
	"errors"
	"fmt"
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
	Log.V(1).Info("Run: " + r.String())
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
		err = &FailedError{
			Command: r.Command(),
			Exit:    cmd.ProcessState.ExitCode(),
			Output:  string(r.Output()),
		}
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

// Command returns the command (path) plus arguments.
func (r *Command) Command() (s string) {
	parts := []string{
		r.Path,
		strings.Join(r.Options, " "),
	}
	s = strings.Join(parts, " ")
	return
}

// String returns a string representation.
func (r *Command) String() (s string) {
	parts := []string{
		"[CMD]",
	}
	if r.Error == nil {
		parts = append(
			parts,
			r.Command(),
			": SUCCEEDED")
	} else {
		parts = append(parts, r.Error.Error())
	}
	s = strings.Join(parts, " ")
	return
}

// FailedError command failed error.
type FailedError struct {
	Command string
	Exit    int
	Output  string
}

func (e *FailedError) Error() (s string) {
	s = fmt.Sprintf(
		"(exit=%d) %s: FAILED, Output: %s",
		e.Exit,
		e.Command,
		e.Output)
	return
}

func (e *FailedError) Is(err error) (matched bool) {
	inst := &FailedError{}
	matched = errors.As(err, &inst)
	return
}
