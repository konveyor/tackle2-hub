package scm

import (
	"os"

	"github.com/konveyor/tackle2-hub/shared/addon/adapter"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/scm"
)

const (
	CREATE = scm.CREATE
)

var (
	addon = adapter.Addon
)

func init() {
	scm.Dir, _ = os.Getwd()
}

type SCM = scm.SCM
type Remote = scm.Remote
type Identity = scm.Identity
type Proxy = scm.Proxy
type ProxyMap = scm.ProxyMap
type Git = scm.Git
type Subversion = scm.Subversion
type Option = scm.Option

// New SCM repository factory.
func New(destDir string, repository api.Repository, identity *api.Identity) (r SCM, err error) {
	insecure, err := scm.Insecure(addon.Client(), repository)
	if err != nil {
		return
	}
	remote := Remote{
		Kind:     repository.Kind,
		URL:      repository.URL,
		Branch:   repository.Branch,
		Path:     repository.Path,
		Insecure: insecure,
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
	r = scm.New(destDir, remote)
	p, err := scm.Proxies(addon.Client())
	if err != nil {
		return
	}
	r.WithProxies(p)
	err = r.Validate()
	return
}
