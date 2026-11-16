package tar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	pathlib "path"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/nas"
)

// Accept determines whether an archive entry may be extracted.
// Returning an error rejects the entry and aborts the extraction.
type Accept func(path string, header *tar.Header) (err error)

// Contained returns an Accept that rejects entries resolving
// outside of the root. Containment is determined lexically and
// does not account for symlinks. Archives are not trusted by
// default. Callers extracting an untrusted archive are expected
// to assign this to Reader.Accept.
func Contained(root string) (accept Accept) {
	root = pathlib.Clean(root)
	accept = func(path string, header *tar.Header) (err error) {
		if path == root || strings.HasPrefix(path, strings.TrimSuffix(root, "/")+"/") {
			return
		}
		err = liberr.New(
			"Archive entry not contained.",
			"entry",
			header.Name,
			"root",
			root)
		return
	}
	return
}

// Reader archive reader.
type Reader struct {
	// Filter (optional) selects the entries to be extracted.
	// Patterns are matched against the entry path resolved within
	// the destination, so the filter Root is expected to be the
	// destination path. Unmatched entries are skipped. When not
	// assigned, all entries are extracted.
	Filter Filter
	// Accept (optional) determines whether an entry may be
	// extracted. When not assigned, all entries are accepted.
	Accept Accept
}

// NewReader returns a new reader.
func NewReader() (reader *Reader) {
	reader = &Reader{}
	return
}

// Extract archive content to the destination path.
func (r *Reader) Extract(outDir string, reader io.Reader) (err error) {
	zipReader, err := gzip.NewReader(reader)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = zipReader.Close()
	}()
	err = os.MkdirAll(outDir, 0777)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	tarReader := tar.NewReader(zipReader)
	for {
		header, nErr := tarReader.Next()
		if nErr != nil {
			if nErr == io.EOF {
				break
			} else {
				err = liberr.Wrap(nErr)
				return
			}
		}
		path := pathlib.Join(outDir, header.Name)
		if !r.Filter.Match(path) {
			continue
		}
		if r.Accept != nil {
			err = r.Accept(path, header)
			if err != nil {
				return
			}
		}
		switch header.Typeflag {
		case tar.TypeDir:
			err = nas.MkDir(path, os.FileMode(header.Mode))
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		case tar.TypeReg:
			err = nas.MkDir(pathlib.Dir(path), 0777)
			if err != nil {
				return
			}
			file, nErr := os.Create(path)
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			_, err = io.Copy(file, tarReader)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			_ = file.Close()
		}
	}
	return
}
