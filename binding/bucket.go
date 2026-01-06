package binding

import (
	pathlib "path"

	api2 "github.com/konveyor/tackle2-hub/api"
)

// Bucket API.
type Bucket struct {
	client *Client
}

// Create a Bucket.
func (h *Bucket) Create(r *api2.Bucket) (err error) {
	err = h.client.Post(api2.BucketsRoute, &r)
	return
}

// Get a bucket.
func (h *Bucket) Get(id uint) (r *api2.Bucket, err error) {
	r = &api2.Bucket{}
	path := Path(api2.BucketRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List buckets.
func (h *Bucket) List() (list []api2.Bucket, err error) {
	list = []api2.Bucket{}
	err = h.client.Get(api2.BucketsRoute, &list)
	return
}

// Delete a bucket.
func (h *Bucket) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.BucketRoute).Inject(Params{api2.ID: id}))
	return
}

// Content returns content API.
func (h *Bucket) Content(id uint) (b *BucketContent) {
	params := Params{
		api2.Wildcard: "",
		api2.ID:       id,
	}
	path := Path(api2.BucketRoute).Inject(params)
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
