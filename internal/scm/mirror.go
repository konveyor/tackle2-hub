package scm

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"github.com/konveyor/tackle2-hub/shared/scm"
	"gorm.io/gorm"
)

var (
	mirrorMap = MirrorMap{content: make(map[string]*Mirror)}
)

// GetMirror returns a mirror for the remote.
func GetMirror(db *gorm.DB, remote scm.Remote) (mirror *Mirror) {
	mirror = mirrorMap.Find(db, remote)
	return
}

// MirrorMap contains a map of mirrors keyed by kind and remote.URL.
type MirrorMap struct {
	content map[string]*Mirror
	mutex   sync.Mutex
}

// Find returns a mirror for the remote.
func (m *MirrorMap) Find(db *gorm.DB, remote scm.Remote) (mirror *Mirror) {
	mirror = func() (mirror *Mirror) {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		mirror, found := m.content[remote.URL]
		if !found {
			mirror = &Mirror{
				Remote: remote,
				DB:     db,
			}
			m.content[remote.URL] = mirror
		}
		return
	}()
	mirror.mutex.Lock()
	defer mirror.mutex.Unlock()
	mirror.Remote.Branch = remote.Branch
	return
}

// Mirror provides a (mirror) repository.
type Mirror struct {
	DB     *gorm.DB
	Remote scm.Remote
	mutex  sync.Mutex
}

// Update the mirror.
func (m *Mirror) Update() (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	err = m.update()
	return
}

// CopyTo updates the mirror and copies the repository to the destination.
func (m *Mirror) CopyTo(path, destDir string) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	err = m.update()
	if err != nil {
		return
	}
	home := m.home()
	if path != "" {
		err = nas.CpDir(filepath.Join(home, path), destDir)
	} else {
		err = nas.CpDir(home, destDir)
	}
	return
}

// update the mirror.
func (m *Mirror) update() (err error) {
	path := m.home()
	err = nas.MkDir(filepath.Dir(path), 0755)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	identity, err := m.identity()
	if err != nil {
		return
	}
	remote := scm.Remote{
		URL:      m.Remote.URL,
		Identity: identity,
	}
	var r scm.SCM
	r, err = New(m.DB, m.home(), remote)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = nas.RmDir(path)
			_ = r.Clean()
		}
	}()
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = r.Fetch()
		} else {
			err = nil
		}
	}
	if err != nil {
		return
	}
	err = r.Update()
	if err != nil {
		return
	}
	err = r.Branch(m.Remote.Branch)
	if err != nil {
		return
	}

	return
}

// identity returns the default source identity.
func (m *Mirror) identity() (id *scm.Identity, err error) {
	md := &model.Identity{}
	db := m.DB
	db = db.Where("kind", "source")
	db = db.Where("default", true)
	err = db.First(md).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = liberr.Wrap(err)
		} else {
			err = nil
		}
		return
	}
	err = secret.Decrypt(md)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	id = &scm.Identity{}
	id.ID = md.ID
	id.Name = md.Name
	id.User = md.User
	id.Password = md.Password
	id.Key = md.Key
	return
}

// home returns the path to the repository.
func (m *Mirror) home() (p string) {
	p = filepath.Join(Home, ".mirror", m.digest())
	return
}

// digest calculates the digest of the mirror based
// on the remote kind and URL.
func (m *Mirror) digest() (d string) {
	d = m.Remote.Digest()
	return
}
