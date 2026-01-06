package scm

import (
	"os"
	"path"

	"github.com/konveyor/tackle2-hub/addon/adapter"
	"github.com/konveyor/tackle2-hub/api"
	scm2 "github.com/konveyor/tackle2-hub/scm"
)

var (
	Dir   = ""
	addon = adapter.Addon
)

func init() {
	Dir, _ = os.Getwd()
}

// New SCM repository factory.
func New(destDir string, repository api.Repository, identity *api.Identity) (r scm2.SCM, err error) {
	remote := scm2.Remote{
		Kind:   repository.Kind,
		URL:    repository.URL,
		Branch: repository.Branch,
		Path:   repository.Path,
	}
	if identity != nil {
		remote.Identity = &scm2.Identity{
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
		svn := &scm2.Subversion{}
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
		git := &scm2.Git{}
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
func proxyMap() (pm scm2.ProxyMap, err error) {
	pm = make(scm2.ProxyMap)
	list, err := addon.Proxy.List()
	if err != nil {
		return
	}
	for _, p := range list {
		if !p.Enabled {
			continue
		}
		proxy := scm2.Proxy{
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
			proxy.Identity = &scm2.Identity{
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
