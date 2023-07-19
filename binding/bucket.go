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
	client *Client
}

//
// Create a Bucket.
func (h *Bucket) Create(r *api.Bucket) (err error) {
	err = h.client.Post(api.BucketsRoot, &r)
	return
}

//
// Get a bucket.
func (h *Bucket) Get(id uint) (r *api.Bucket, err error) {
	r = &api.Bucket{}
	path := Path(api.BucketRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List buckets.
func (h *Bucket) List() (list []api.Bucket, err error) {
	list = []api.Bucket{}
	err = h.client.Get(api.BucketsRoot, &list)
	return
}

//
// Delete a bucket.
func (h *Bucket) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.BucketRoot).Inject(Params{api.ID: id}))
	return
}

//
// Content returns content API.
func (h *Bucket) Content(id uint) (b *BucketContent) {
	params := Params{
		api.Wildcard: "",
		api.ID:       id,
	}
	path := Path(api.BucketRoot).Inject(params)
	b = &BucketContent{
		root:   path,
		client: h.client,
	}
	return
}

//
// BucketContent API.
type BucketContent struct {
	client *Client
	root   string
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
// The source (root) is relative to the bucket root.
func (h *BucketContent) Get(source, destination string) (err error) {
	err = h.client.BucketGet(pathlib.Join(h.root, source), destination)
	return
}

//
// Put writes to the bucket.
// The destination (root) is relative to the bucket root.
func (h *BucketContent) Put(source, destination string) (err error) {
	err = h.client.BucketPut(source, pathlib.Join(h.root, destination))
	return
}

//
// Delete deletes content at the specified root.
// The source is relative to the bucket root.
func (h *BucketContent) Delete(path string) (err error) {
	err = h.client.Delete(pathlib.Join(h.root, path))
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
