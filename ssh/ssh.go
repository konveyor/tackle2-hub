/*
Package ssh provides a SSH related functionality.
*/
package ssh

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/command"
	"github.com/konveyor/tackle2-hub/nas"
)

var (
	Log        = logr.WithName("ssh")
	NewCommand func(string) *command.Command
	Home       = ""
)

func init() {
	Home, _ = os.Getwd()
	NewCommand = command.New
}

// Key is and SSH key.
type Key struct {
	ID         uint
	Name       string
	Content    string
	Passphrase string
}

// Agent agent.
type Agent struct {
}

// Start the ssh-agent.
func (r *Agent) Start() (err error) {
	pid := os.Getpid()
	socket := fmt.Sprintf("/tmp/agent.%d", pid)
	cmd := NewCommand("/usr/bin/ssh-agent")
	cmd.Env = append(os.Environ(), "HOME="+r.home())
	cmd.Options.Add("-a", socket)
	err = cmd.Run()
	if err != nil {
		return
	}
	_ = os.Setenv("SSH_AUTH_SOCK", socket)
	err = nas.MkDir(r.sshDir(), 0700)
	if err != nil {
		return
	}

	Log.Info("[SSH] Agent started.", "home", r.home())

	return
}

// Add an ssh key.
func (r *Agent) Add(key Key, host string) (err error) {
	if key.Content == "" {
		return
	}
	Log.Info("[SSH] Adding key: %s" + key.Name)
	suffix := fmt.Sprintf("id_%d", key.ID)
	path := filepath.Join(
		r.sshDir(),
		suffix)
	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		0600)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	_, err = f.Write([]byte(r.format(key.Content)))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	}
	_ = f.Close()
	ask, err := r.writeAsk(key)
	if err != nil {
		return
	}
	ctx, fn := context.WithTimeout(context.TODO(), 3*time.Second)
	defer fn()
	cmd := NewCommand("/usr/bin/ssh-add")
	cmd.Env = append(
		os.Environ(),
		"DISPLAY=:0",
		"SSH_ASKPASS="+ask,
		"HOME="+r.home())
	cmd.Options.Add(path)
	err = cmd.RunWith(ctx)
	if err != nil {
		return
	}
	Log.Info("[SSH] Created: " + path)
	return
}

// Ensure key formatting.
func (r *Agent) format(in string) (out string) {
	if in != "" {
		out = strings.TrimSpace(in) + "\n"
	}
	return
}

// writeAsk writes script that returns the key password.
func (r *Agent) writeAsk(key Key) (path string, err error) {
	f, err := os.CreateTemp("", "askpass-*.sh")
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	path = f.Name()
	script := "#!/bin/sh\n"
	script += "echo " + key.Passphrase
	_ = os.Chmod(path, 0700)
	_, err = f.Write([]byte(script))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	}
	return
}

// home returns the path to the client home directory.
func (r *Agent) home() string {
	return Home
}

// sshDir returns the directory where client keys are stored.
func (r *Agent) sshDir() (p string) {
	p = filepath.Join(r.home(), ".ssh")
	return
}
