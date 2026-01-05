package scm

import (
	"os"
	"path/filepath"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/scm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	Home = ""
)

func init() {
	Home, _ = os.Getwd()
}

type SCM = scm.SCM
type Remote = scm.Remote
type Identity = scm.Identity
type Proxy = scm.Proxy
type ProxyMap = scm.ProxyMap

// New SCM repository factory.
func New(db *gorm.DB, destDir string, remote Remote) (r SCM, err error) {
	switch remote.Kind {
	case "subversion":
		m := model.Setting{}
		err = db.First(&m, "key", "svn.insecure.enabled").Error
		if err != nil {
			return
		}
		err = m.As(&remote.Insecure)
		if err != nil {
			return
		}
		svn := &scm.Subversion{}
		svn.Remote = remote
		svn.Path = destDir
		svn.Home = filepath.Join(Home, ".svn", svn.Id())
		svn.Proxies, err = proxyMap(db)
		if err != nil {
			return
		}
		r = svn
	default:
		m := model.Setting{}
		err = db.First(&m, "key", "git.insecure.enabled").Error
		if err != nil {
			return
		}
		err = m.As(&remote.Insecure)
		if err != nil {
			return
		}
		git := &scm.Git{}
		git.Remote = remote
		git.Path = destDir
		git.Home = filepath.Join(Home, ".git", git.Id())
		git.Proxies, err = proxyMap(db)
		if err != nil {
			return
		}
		r = git
	}
	err = r.Validate()
	if err != nil {
		return
	}
	return
}

// proxyMap returns a map of proxies.
func proxyMap(db *gorm.DB) (pm ProxyMap, err error) {
	pm = make(ProxyMap)
	var list []model.Proxy
	db = db.Preload(clause.Associations)
	err = db.Find(&list).Error
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
			proxy.Identity = &Identity{
				ID:       p.Identity.ID,
				Name:     p.Identity.Name,
				User:     p.Identity.User,
				Password: p.Identity.Password,
				Key:      p.Identity.Key,
			}
		}
		pm[p.Kind] = proxy
	}
	return
}
