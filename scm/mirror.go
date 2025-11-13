package scm

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sync"

	"github.com/konveyor/tackle2-hub/nas"
	"gorm.io/gorm"
)

var (
	mirrorMap = MirrorMap{content: make(map[string]*Mirror)}
)

// GetMirror returns a mirror for the remote.
func GetMirror(db *gorm.DB, remote Remote) (mirror *Mirror) {
	mirror = mirrorMap.Find(db, remote)
	return
}

// MirrorMap contains a map of mirrors keyed by kind and remote.URL.
type MirrorMap struct {
	content map[string]*Mirror
	mutex   sync.Mutex
}

// Find returns a mirror for the remote.
func (m *MirrorMap) Find(db *gorm.DB, remote Remote) (mirror *Mirror) {
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
}

// Mirror provides a (mirror) repository.
type Mirror struct {
	DB     *gorm.DB
	Remote Remote
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
func (m *Mirror) CopyTo(destDir string) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	err = m.update()
	if err != nil {
		return
	}
	home := m.home()
	if m.Remote.Path != "" {
		err = nas.CpDir(filepath.Join(home, m.Remote.Path), destDir)
	} else {
		err = nas.CpDir(home, destDir)
	}
	return
}

// update the mirror.
func (m *Mirror) update() (err error) {
	path := m.home()
	remote := &Remote{
		URL: m.Remote.URL,
	}
	var r SCM
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
	if m.Remote.Branch != "" {
		err = r.Branch(m.Remote.Branch)
		if err != nil {
			return
		}
	}
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
	h := fnv.New64a()
	_, _ = h.Write([]byte(m.Remote.Kind))
	_, _ = h.Write([]byte(m.Remote.URL))
	n := h.Sum64()
	d = fmt.Sprintf("%x", n)
	return
}
