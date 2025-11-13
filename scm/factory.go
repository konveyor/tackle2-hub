package scm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"gorm.io/gorm"
)

// New SCM repository factory.
func New(db *gorm.DB, destDir string, remote *Remote, option ...any) (r SCM, err error) {
	switch remote.Kind {
	case "subversion":
		m := model.Setting{}
		err = db.First(&m, "key", "svn.insecure.enabled").Error
		if err != nil {
			return
		}
		svn := &Subversion{}
		svn.Home = filepath.Join(Home, svn.Id())
		svn.Path = destDir
		svn.Remote = *remote
		svn.Proxies, err = proxyMap(db)
		if err != nil {
			return
		}
		err = m.As(&svn.Insecure)
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
		git := &Git{}
		git.Home = filepath.Join(Home, git.Id())
		git.Path = destDir
		git.Remote = *remote
		git.Proxies, err = proxyMap(db)
		if err != nil {
			return
		}
		err = m.As(&git.Insecure)
		if err != nil {
			return
		}
		r = git
	}
	err = r.Validate()
	if err != nil {
		return
	}
	for _, opt := range option {
		err = r.Use(opt)
		if err != nil {
			return
		}
	}
	return
}

type Factory struct {
}

// proxyMap returns a map of proxies.
func proxyMap(db *gorm.DB) (mp map[string]Proxy, err error) {
	var list []model.Proxy
	err = db.Find(&list).Error
	if err != nil {
		return
	}
	for _, p := range list {
		if !p.Enabled {
			continue
		}
		mp[p.Kind] = Proxy{
			ID:       p.ID,
			Kind:     p.Kind,
			Host:     p.Host,
			Port:     p.Port,
			Excluded: p.Excluded,
		}
	}
	return
}

// Base SCM.
type Base struct {
	Authenticated
	Home    string
	Proxies map[string]Proxy
	Remote  Remote
	Path    string
	//
	id string
}

// Id returns the unique id.
func (b *Base) Id() string {
	if b.id == "" {
		b.id = uuid.New().String()
	}
	return b.id
}

// Clean deletes created files.
func (b *Base) Clean() (err error) {
	err = nas.RmDir(b.Home)
	return
}

// Validate the repository.
// Ensures that Home and Path either:
// - do not exist.
// - are empty directories.
func (b *Base) Validate() (err error) {
	err = b.mustEmptyDir(b.Home)
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
