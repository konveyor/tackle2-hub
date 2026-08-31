/*
Package scm provides objects for working with
SCM (Software Configuration Management) repositories.
*/
package scm

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
)

var (
	Log = logr.New("SCM", 0)
	Dir = ""
)

func init() {
	Dir, _ = os.Getwd()
}

// SCM interface.
type SCM interface {
	Id() string
	Validate() (err error)
	WithProxies(p ProxyMap)
	WithIdentity(id *Identity)
	WithInsecure(enabled bool)
	Fetch() (err error)
	Update() (err error)
	Branch(ref string, options ...Option) (err error)
	Commit(files []string, msg string) (err error)
	Push() (err error)
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

// New SCM repository factory.
func New(destDir string, repository api.Repository) (r SCM) {
	remote := Remote{
		Kind:   repository.Kind,
		URL:    repository.URL,
		Branch: repository.Branch,
		Path:   repository.Path,
	}
	switch remote.Kind {
	case "svn",
		"subversion":
		svn := &Subversion{}
		svn.Remote = remote
		svn.Path = destDir
		svn.Home = filepath.Join(Dir, ".svn", svn.Id())
		r = svn
	default:
		git := &Git{}
		git.Remote = remote
		git.Path = destDir
		git.Home = filepath.Join(Dir, ".git", git.Id())
		r = git
	}
	return
}

// Insecure returns the hub insecure settings.
func Insecure(client *binding.RichClient, r api.Repository) (enabled bool, err error) {
	var key string
	switch r.Kind {
	case "svn",
		"subversion":
		key = "svn.insecure.enabled"
	default:
		key = "git.insecure.enabled"
	}
	enabled, err = client.Setting.Bool(key)
	return
}

// Proxies returns a map of hub proxies.
func Proxies(client *binding.RichClient) (pm ProxyMap, err error) {
	pm = make(ProxyMap)
	list, err := client.Proxy.List()
	if err != nil {
		return
	}
	for _, p := range list {
		if !p.Enabled {
			continue
		}
		proxy := Proxy{
			ID:       p.ID,
			Kind:     p.Kind,
			Host:     p.Host,
			Port:     p.Port,
			Excluded: p.Excluded,
		}
		if p.Identity != nil {
			var identity *api.Identity
			identity, err = client.Identity.Get(p.Identity.ID)
			if err != nil {
				return
			}
			proxy.Identity = &Identity{
				ID:       identity.ID,
				Name:     identity.Name,
				User:     identity.User,
				Password: identity.Password,
				Key:      identity.Key,
			}
		}
		pm[p.Kind] = proxy
	}
	return
}
