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
	Fetch() (err error)
	Update() (err error)
	Branch(ref string) (err error)
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
func New(
	destDir string,
	repository api.Repository,
	identity *api.Identity,
	client *binding.RichClient) (r SCM, err error) {
	//
	remote := Remote{
		Kind:   repository.Kind,
		URL:    repository.URL,
		Branch: repository.Branch,
		Path:   repository.Path,
	}
	if identity != nil {
		remote.Identity = &Identity{
			ID:       identity.ID,
			Name:     identity.Name,
			User:     identity.User,
			Password: identity.Password,
			Key:      identity.Key,
		}
	}
	switch remote.Kind {
	case "subversion":
		remote.Insecure, err = client.Setting.Bool("svn.insecure.enabled")
		if err != nil {
			return
		}
		svn := &Subversion{}
		svn.Remote = remote
		svn.Path = destDir
		svn.Home = filepath.Join(Dir, ".svn", svn.Id())
		svn.Proxies, err = proxyMap(client)
		if err != nil {
			return
		}
		r = svn
	default:
		remote.Insecure, err = client.Setting.Bool("git.insecure.enabled")
		if err != nil {
			return
		}
		git := &Git{}
		git.Remote = remote
		git.Path = destDir
		git.Home = filepath.Join(Dir, ".git", git.Id())
		git.Proxies, err = proxyMap(client)
		if err != nil {
			return
		}
		r = git
	}
	err = r.Validate()
	return
}

// proxyMap returns a map of proxies.
func proxyMap(client *binding.RichClient) (pm ProxyMap, err error) {
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
