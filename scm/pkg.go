/*
Package scm provides objects for working with
SCM (Software Configuration Management) repositories.
*/
package scm

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/command"
	"github.com/pkg/errors"
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
	Branch(ref string) (err error)
	Commit(files []string, msg string) (err error)
	Head() (commit string, err error)
	Use(option any) (err error)
	Clean()
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
	Kind   string
	URL    string
	Branch string
	Tag    string
	Path   string
}

// Authenticated repository.
type Authenticated struct {
	Identity Identity
	Insecure bool
}

// Use option.
func (a *Authenticated) Use(option any) (err error) {
	switch opt := option.(type) {
	case *Identity:
		if opt != nil {
			a.Identity = *opt
		}
	case Identity:
		a.Identity = opt
	default:
		err = errors.Errorf("Invalid option: %T", opt)
	}
	return
}
