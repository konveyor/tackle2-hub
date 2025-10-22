package scm

import (
	pathlib "path"

	"github.com/google/uuid"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"gorm.io/gorm"
)

// New SCM repository factory.
// Options:
// - *api.Ref
// - api.Ref
// - *api.Identity
// - api.Identity
func New(db *gorm.DB, destDir string, remote *Remote, option ...any) (r SCM, err error) {
	switch remote.Kind {
	case "subversion":
		m := model.Setting{}
		err = db.First(&m, "key", "svn.insecure.enabled").Error
		if err != nil {
			return
		}
		svn := &Subversion{}
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

type Base struct {
	Authenticated
	HomeRoot string
	Proxies  map[string]Proxy
	Remote   Remote
	Path     string
	//
	id string
}

func (b *Base) Id() string {
	if b.id == "" {
		b.id = uuid.New().String()
	}
	return b.id
}

// Home returns the Git home directory path.
func (b *Base) Home() (home string) {
	home = pathlib.Join(
		b.HomeRoot,
		b.Id(),
		".git")
	return
}

func (b *Base) Clean() {
	_ = nas.RmDir(b.Home())
	_ = nas.RmDir(b.Path)
}
