/*
Package scm provides objects for working with
SCM (Software Configuration Management) repositories.
*/
package scm

import (
	"fmt"
	"hash/fnv"

	"github.com/jortel/go-utils/logr"
)

var (
	Log = logr.New("SCM", 0)
)

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

// ProxyMap keyed by scheme.
type ProxyMap map[string]Proxy

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

// Digest calculates the digest of the remote based
// on the remote kind and URL.
func (r *Remote) Digest() (d string) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(r.Kind))
	_, _ = h.Write([]byte(r.URL))
	n := h.Sum64()
	d = fmt.Sprintf("%x", n)
	return
}
