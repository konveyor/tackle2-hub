package binding

import (
	"io"
	pathlib "path"

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
