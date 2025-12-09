package tar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	pathlib "path"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/nas"
)

// NewReader returns a new reader.
func NewReader() (reader *Reader) {
	reader = &Reader{}
	return
}

// Reader archive reader.
type Reader struct {
	Filter Filter
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
			return
		}
		switch header.Typeflag {
		case tar.TypeDir:
			err = nas.MkDir(path, os.FileMode(header.Mode))
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		case tar.TypeReg:
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
