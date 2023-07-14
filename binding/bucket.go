package binding

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	pathlib "path"
	"path/filepath"

	"github.com/konveyor/tackle2-hub/api"
)

//
// Bucket API.
type Bucket struct {
	// hub API client.
	Client *Client
	// root path
	path string
}

//
// Create a Bucket.
func (h *Bucket) Create(r *api.Bucket) (err error) {
	err = h.Client.Post(api.BucketsRoot, &r)
	return
}

//
// List Buckets.
func (h *Bucket) List() (list []api.Bucket, err error) {
	list = []api.Bucket{}
	err = h.Client.Get(api.BucketsRoot, &list)
	return
}

//
// Get reads from the bucket.
// The source (path) is relative to the bucket root.
func (h *Bucket) Get(source, destination string) (err error) {
	err = h.Client.BucketGet(pathlib.Join(h.path, source), destination)
	return
}

//
// Put writes to the bucket.
// The destination (path) is relative to the bucket root.
func (h *Bucket) Put(source, destination string) (err error) {
	err = h.Client.BucketPut(source, pathlib.Join(h.path, destination))
	return
}

//
// Delete deletes content at the specified path.
// The path is relative to the bucket root.
func (h *Bucket) Delete(path string) (err error) {
	err = h.Client.Delete(pathlib.Join(h.path, path))
	return
}

func (h *Bucket) GetDir(file io.Reader, path string) (err error) {
	err = h.Client.getDir(file, path)
	return
}

func (h *Bucket) PutDir(file io.Writer, path string) (err error) {
	err = h.Client.putDir(file, path)
	return
}

func (h *Bucket) Compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		header.Name = filepath.ToSlash(file)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}
