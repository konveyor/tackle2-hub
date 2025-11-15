/*
Package scm provides objects for working with
SCM (Software Configuration Management) repositories.
*/
package scm

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/command"
)

var (
	Log        = logr.WithName("SCM")
	NewCommand func(string) *command.Command
)

func init() {
	NewCommand = command.New
}

// SCM interface.
type SCM interface {
	Id() string
	Validate() (err error)
	Fetch() (err error)
	Update() (err error)
	Branch(ref string) (err error)
	Commit(files []string, msg string) (err error)
	Head() (commit string, err error)
	Clean() (err error)
}

// Proxy defines a proxy.
type Proxy struct {
	ID       uint
	Kind     string
	Host     string
	Port     int
	Excluded []string
	Identity *Identity
}

// Identity defines an identity.
type Identity struct {
	ID       uint
	Name     string
	User     string `json:"user"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

// Remote defines an SCM remote.
type Remote struct {
	Kind     string
	URL      string
	Branch   string
	Path     string
	Identity *Identity
	Insecure bool
}
