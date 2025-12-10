package scm

import (
	"os"
	"path"

	"github.com/konveyor/tackle2-hub/shared/addon/adapter"
	"github.com/konveyor/tackle2-hub/shared/addon/command"
	"github.com/konveyor/tackle2-hub/shared/addon/sink"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/scm"
)

var (
	Dir   = ""
	addon = adapter.Addon
)

func init() {
	Dir, _ = os.Getwd()
	scm.NewCommand = command.New
	scm.Log = scm.Log.WithSink(sink.New(true))
}

type SCM = scm.SCM
type Remote = scm.Remote
type Identity = scm.Identity
type ProxyMap = scm.ProxyMap
type Proxy = scm.Proxy
type Subversion = scm.Subversion
type Git = scm.Git

func init() {
	scm.NewCommand = command.New
}

// New SCM repository factory.
func New(destDir string, repository api.Repository, identity *api.Identity) (r SCM, err error) {
	remote := scm.Remote{
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
		remote.Insecure, err = addon.Setting.Bool("svn.insecure.enabled")
		if err != nil {
			return
		}
		svn := &Subversion{}
		svn.Remote = remote
		svn.Path = destDir
		svn.Home = path.Join(Dir, ".svn", svn.Id())
		svn.Proxies, err = proxyMap()
		if err != nil {
			return
		}
		r = svn
	default:
		remote.Insecure, err = addon.Setting.Bool("git.insecure.enabled")
		if err != nil {
			return
		}
		git := &Git{}
		git.Remote = remote
		git.Path = destDir
		git.Home = path.Join(Dir, ".git", git.Id())
		git.Proxies, err = proxyMap()
		if err != nil {
			return
		}
		r = git
	}
	err = r.Validate()
	return
}

// proxyMap returns a map of proxies.
func proxyMap() (pm ProxyMap, err error) {
	pm = make(ProxyMap)
	list, err := addon.Proxy.List()
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
			identity, err = addon.Identity.Get(p.Identity.ID)
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
