package tar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	liberr "github.com/jortel/go-utils/error"
)

// NewWriter returns a new writer.
func NewWriter(output io.Writer) (writer *Writer) {
	writer = &Writer{}
	writer.Open(output)
	runtime.SetFinalizer(
		writer,
		func(r *Writer) {
			r.Close()
		})
	return
}

// Writer is a Zipped TAR streamed writer.
type Writer struct {
	Filter Filter
	//
	drained   chan int
	tarWriter *tar.Writer
	bridge    struct {
		reader *io.PipeReader
		writer *io.PipeWriter
	}
}

// Open the writer.
func (r *Writer) Open(output io.Writer) {
	if r.tarWriter != nil {
		return
	}
	r.drained = make(chan int)
	r.bridge.reader, r.bridge.writer = io.Pipe()
	r.tarWriter = tar.NewWriter(r.bridge.writer)
	zipWriter := gzip.NewWriter(output)
	go func() {
		defer func() {
			_ = zipWriter.Close()
			r.drained <- 0
		}()
		_, _ = io.Copy(zipWriter, r.bridge.reader)
	}()
}

// AssertDir validates the path is a readable directory.
func (r *Writer) AssertDir(pathIn string) (err error) {
	st, err := os.Stat(pathIn)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if !st.IsDir() {
		err = liberr.New("Directory path expected.")
		return
	}
	err = filepath.Walk(
		pathIn,
		func(path string, info os.FileInfo, nErr error) (err error) {
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			if path == pathIn {
				return
			}
			if !r.Filter.Match(path) {
				return
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			switch header.Typeflag {
			case tar.TypeReg:
				f, nErr := os.Open(path)
				if nErr != nil {
					err = liberr.Wrap(nErr)
					return
				}
				_ = f.Close()
			}
			return
		})
	return
}

// AddDir adds a directory.
func (r *Writer) AddDir(pathIn string) (err error) {
	if r.tarWriter == nil {
		err = liberr.New("Writer not open.")
		return
	}
	err = r.AssertDir(pathIn)
	if err != nil {
		return
	}
	err = filepath.Walk(
		pathIn,
		func(path string, info os.FileInfo, nErr error) (err error) {
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			if path == pathIn {
				return
			}
			if !r.Filter.Match(path) {
				return
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			header.Name = strings.Replace(path, pathIn, "", 1)
			switch header.Typeflag {
			case tar.TypeDir:
				err = r.tarWriter.WriteHeader(header)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
			case tar.TypeReg:
				err = r.tarWriter.WriteHeader(header)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				file, nErr := os.Open(path)
				if err != nil {
					err = liberr.Wrap(nErr)
					return
				}
				defer func() {
					_ = file.Close()
				}()
				_, err = io.Copy(r.tarWriter, file)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
			}
			return
		})
	return
}

// AssertFile validates the path is a readable file.
func (r *Writer) AssertFile(pathIn string) (err error) {
	st, err := os.Stat(pathIn)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if st.IsDir() {
		err = liberr.New("File path expected.")
		return
	}
	f, nErr := os.Open(pathIn)
	if nErr != nil {
		err = liberr.Wrap(nErr)
		return
	}
	_ = f.Close()
	return
}

// AddFile adds a file.
func (r *Writer) AddFile(pathIn, destPath string) (err error) {
	if r.tarWriter == nil {
		err = liberr.New("Writer not open.")
		return
	}
	err = r.AssertFile(pathIn)
	if err != nil {
		return
	}
	st, err := os.Stat(pathIn)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	header, err := tar.FileInfoHeader(st, "")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	header.Name = destPath
	err = r.tarWriter.WriteHeader(header)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	file, nErr := os.Open(pathIn)
	if err != nil {
		err = liberr.Wrap(nErr)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = io.Copy(r.tarWriter, file)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Close the writer.
func (r *Writer) Close() {
	if r.tarWriter == nil {
		return
	}
	_ = r.tarWriter.Close()
	_ = r.bridge.writer.Close()
	_ = <-r.drained
}
