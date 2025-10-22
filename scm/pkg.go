package scm

import (
	"fmt"
	"strings"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/pkg/errors"
)

// Logf logger.
var Logf func(s string, v ...any)

func init() {
	Logf = func(s string, v ...any) {
		fmt.Printf(s, v...)
		fmt.Print("\n")
	}
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

// Remote defines and SCM.
type Remote api.Repository

func (r *Remote) String() (s string) {
	return fmt.Sprintf(
		"[%s]URL:%s",
		strings.ToUpper(r.Kind),
		r.URL)
}

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

// Insecure option
type Insecure bool

// Use option.
// Options:
// - Insecure
// - *api.Identity
// - api.Identity
func (a *Authenticated) Use(option any) (err error) {
	switch opt := option.(type) {
	case Insecure:
		a.Insecure = bool(opt)
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
