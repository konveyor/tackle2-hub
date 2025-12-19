/*
Package ssh provides a SSH related functionality.
*/
package ssh

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/command"
	"github.com/konveyor/tackle2-hub/shared/nas"
)

var (
	Log   = logr.New("ssh", 0)
	Home  = ""
	agent = Agent{}
)

func init() {
	Home, _ = os.Getwd()
	err := agent.Start()
	if err != nil {
		panic(err)
	}
}

// Agent agent.
type Agent struct {
}

// Start the ssh-agent.
func (r *Agent) Start() (err error) {
	err = nas.MkDir(r.sshDir(), 0700)
	if err != nil {
		return
	}
	pid := os.Getpid()
	socket := fmt.Sprintf("/tmp/agent.%d", pid)
	cmd := command.New("/usr/bin/ssh-agent")
	cmd.Env = append(os.Environ(), "HOME="+r.home())
	cmd.Options.Add("-a", socket)
	err = cmd.Run()
	if err != nil {
		return
	}
	_ = os.Setenv("SSH_AUTH_SOCK", socket)
	Log.V(1).Info("[SSH] Agent started.", "home", r.home())
	return
}

// Add an ssh key.
func (r *Agent) Add(key Key) (err error) {
	if key.Content == "" {
		return
	}
	keyPath, err := r.writeKey(key)
	if err != nil {
		return
	}
	askPath, err := r.writeAsk(key)
	if err != nil {
		return
	}
	ctx, fn := context.WithTimeout(context.TODO(), 3*time.Second)
	defer fn()
	cmd := command.New("/usr/bin/ssh-add")
	cmd.Env = append(
		os.Environ(),
		"DISPLAY=:0",
		"SSH_ASKPASS="+askPath,
		"HOME="+r.home())
	cmd.Options.Add(keyPath)
	err = cmd.RunWith(ctx)
	if err != nil {
		_ = os.Remove(keyPath)
		_ = os.Remove(askPath)
		return
	}
	Log.V(1).Info("[SSH] Created: " + keyPath)
	return
}

// writeKey writes the ssh key.
func (r *Agent) writeKey(key Key) (path string, err error) {
	suffix := fmt.Sprintf("id_%d", key.ID)
	path = filepath.Join(
		r.sshDir(),
		suffix)
	err = r.fileWrite(path, key.Formatted(), 0600)
	return
}

// writeAsk writes script that returns the key password.
func (r *Agent) writeAsk(key Key) (path string, err error) {
	suffix := fmt.Sprintf("%d_askpass.sh", key.ID)
	path = filepath.Join(
		r.sshDir(),
		suffix)
	script := "#!/bin/sh\n"
	script += "echo " + key.Passphrase
	err = r.fileWrite(path, script, 0700)
	return
}

// fileWrite provides an atomic file create/update.
func (r *Agent) fileWrite(path string, content string, mode os.FileMode) (err error) {
	f, err := os.CreateTemp(filepath.Dir(path), "")
	if err != nil {
		err = liberr.Wrap(err, "path", path)
		return
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	err = os.Chmod(f.Name(), mode)
	if err != nil {
		err = liberr.Wrap(err, "path", path)
		return
	}
	_, err = f.Write([]byte(content))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = f.Sync()
	if err != nil {
		err = liberr.Wrap(err, "path", path)
		return
	}
	_ = f.Close()
	err = os.Rename(f.Name(), path)
	if err != nil {
		err = liberr.Wrap(err, "path", path)
		return
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
