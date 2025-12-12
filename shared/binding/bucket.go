package binding

import (
	pathlib "path"

	"github.com/konveyor/tackle2-hub/shared/api"
)

// Bucket API.
type Bucket struct {
	client *Client
}

// Create a Bucket.
func (h *Bucket) Create(r *api.Bucket) (err error) {
	err = h.client.Post(api.BucketsRoute, &r)
	return
}

// Get a bucket.
func (h *Bucket) Get(id uint) (r *api.Bucket, err error) {
	r = &api.Bucket{}
	path := Path(api.BucketRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List buckets.
func (h *Bucket) List() (list []api.Bucket, err error) {
	list = []api.Bucket{}
	err = h.client.Get(api.BucketsRoute, &list)
	return
}

// Delete a bucket.
func (h *Bucket) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.BucketRoute).Inject(Params{api.ID: id}))
	return
}

// Content returns content API.
func (h *Bucket) Content(id uint) (b *BucketContent) {
	params := Params{
		api.Wildcard: "",
		api.ID:       id,
	}
	path := Path(api.BucketRoute).Inject(params)
	b = &BucketContent{
		root:   path,
		client: h.client,
	}
	return
}

// BucketContent API.
type BucketContent struct {
	client *Client
	root   string
}

// Get reads from the bucket.
// The source (root) is relative to the bucket root.
func (h *BucketContent) Get(source, destination string) (err error) {
	err = h.client.BucketGet(pathlib.Join(h.root, source), destination)
	return
}

// Put writes to the bucket.
// The destination (root) is relative to the bucket root.
func (h *BucketContent) Put(source, destination string) (err error) {
	err = h.client.BucketPut(source, pathlib.Join(h.root, destination))
	return
}

// Delete deletes content at the specified root.
// The source is relative to the bucket root.
func (h *BucketContent) Delete(path string) (err error) {
	err = h.client.Delete(pathlib.Join(h.root, path))
	return
}
