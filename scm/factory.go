package scm

import (
	"path/filepath"

	"github.com/google/uuid"
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
func (b *Base) Clean() {
	_ = nas.RmDir(b.Home)
	_ = nas.RmDir(b.Path)
}
