/*
Package scm provides objects for working with
SCM (Software Configuration Management) repositories.
*/
package scm

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/api"
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

// Remote defines an SCM remote..
type Remote = api.Repository

// Proxy defines a web proxy.
type Proxy struct {
	api.Proxy
	Identity *api.Identity
}

// Authenticated repository.
type Authenticated struct {
	Identity api.Identity
	Insecure bool
}

// Use option.
func (a *Authenticated) Use(option any) (err error) {
	switch opt := option.(type) {
	case *api.Identity:
		if opt != nil {
			a.Identity = *opt
		}
	case api.Identity:
		a.Identity = opt
	default:
		err = errors.Errorf("Invalid option: %T", opt)
	}
	return
}
