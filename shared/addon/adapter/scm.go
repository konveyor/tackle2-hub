package adapter

import (
	"os"
	"path"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/scm"
)

var (
	Dir = ""
)

func init() {
	Dir, _ = os.Getwd()
}

type SCM struct {
}

// New SCM repository factory.
func (_ *SCM) New(destDir string, repository api.Repository, identity *api.Identity) (r scm.SCM, err error) {
	remote := scm.Remote{
		Kind:   repository.Kind,
		URL:    repository.URL,
		Branch: repository.Branch,
		Path:   repository.Path,
	}
	if identity != nil {
		remote.Identity = &scm.Identity{
			ID:       identity.ID,
			Name:     identity.Name,
			User:     identity.User,
			Password: identity.Password,
			Key:      identity.Key,
		}
	}
	switch remote.Kind {
	case "subversion":
		remote.Insecure, err = Addon.Setting.Bool("svn.insecure.enabled")
		if err != nil {
			return
		}
		svn := &scm.Subversion{}
		svn.Remote = remote
		svn.Path = destDir
		svn.Home = path.Join(Dir, ".svn", svn.Id())
		svn.Proxies, err = proxyMap()
		if err != nil {
			return
		}
		r = svn
	default:
		remote.Insecure, err = Addon.Setting.Bool("git.insecure.enabled")
		if err != nil {
			return
		}
		git := &scm.Git{}
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
func proxyMap() (pm scm.ProxyMap, err error) {
	pm = make(scm.ProxyMap)
	list, err := Addon.Proxy.List()
	if err != nil {
		return
	}
	for _, p := range list {
		if !p.Enabled {
			continue
		}
		proxy := scm.Proxy{
			ID:       p.ID,
			Kind:     p.Kind,
			Host:     p.Host,
			Port:     p.Port,
			Excluded: p.Excluded,
		}
		if p.Identity != nil {
			var identity *api.Identity
			identity, err = Addon.Identity.Get(p.Identity.ID)
			if err != nil {
				return
			}
			proxy.Identity = &scm.Identity{
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
