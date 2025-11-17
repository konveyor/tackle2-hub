package scm

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"gorm.io/gorm"
)

// New SCM repository factory.
func New(db *gorm.DB, destDir string, remote *Remote) (r SCM, err error) {
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
		svn := &Subversion{}
		svn.Remote = *remote
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
		git := &Git{}
		git.Remote = *remote
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

// Base SCM.
type Base struct {
	Home    string
	Proxies map[string]Proxy
	Remote  Remote
	Path    string
	//
	id string
}

// Id returns the unique id.
// Based on the remote digest and the (LOCAL) path.
func (b *Base) Id() string {
	if b.id == "" {
		h := fnv.New64a()
		_, _ = h.Write([]byte(b.Remote.digest()))
		_, _ = h.Write([]byte(b.Path))
		n := h.Sum64()
		b.id = fmt.Sprintf("%x", n)
	}
	return b.id
}

// Clean deletes created files.
func (b *Base) Clean() (err error) {
	err = nas.RmDir(b.Home)
	return
}

// mustEmptyDir ensures the path either:
// - does not exist.
// - is an empty directory.
func (b *Base) mustEmptyDir(p string) (err error) {
	defer func() {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	}()
	st, err := os.Stat(p)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if !st.IsDir() {
		err = fmt.Errorf("%s: must be a directory", p)
		return
	}
	entries, err := os.ReadDir(p)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if len(entries) > 0 {
		err = fmt.Errorf("%s: must be empty", p)
		return
	}
	return
}
