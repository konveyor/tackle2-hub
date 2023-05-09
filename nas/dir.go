/*
Package nas provides support for efficiently working with
network attached storage (NAS).
*/
package nas

import (
	"errors"
	liberr "github.com/jortel/go-utils/error"
	"os"
	"os/exec"
)

//
// RmDir deletes the directory.
func RmDir(path string) (err error) {
	cmd := exec.Command("/usr/bin/rm", "-rf", path)
	err = cmd.Run()
	return
}

//
// HasDir return if the path exists.
func HasDir(path string) (found bool, err error) {
	found, err = Exists(path)
	return
}

//
// MkDir ensures the directory exists.
func MkDir(path string, mode os.FileMode) (err error) {
	err = os.MkdirAll(path, mode)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			err = nil
		} else {
			err = liberr.Wrap(
				err,
				"path",
				path)
		}
	}
	return
}

//
// Exists return if the path exists.
func Exists(path string) (found bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		found = true
		return
	}
	if !os.IsNotExist(err) {
		err = liberr.Wrap(
			err,
			"path",
			path)
	} else {
		err = nil
	}
	return
}
