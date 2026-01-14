package scm

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/nas"
)

var Home = ""

func init() {
	Home, _ = os.Getwd()
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
		_, _ = h.Write([]byte(b.Remote.Digest()))
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
