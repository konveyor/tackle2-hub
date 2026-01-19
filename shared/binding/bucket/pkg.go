package bucket

import (
	pathlib "path"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client *client.Client) (h Bucket) {
	h = Bucket{client: client}
	return
}

func NewContent(client *client.Client, root string) (h Content) {
	h = Content{
		client: client,
		root:   root,
	}
	return
}

// Bucket API.
type Bucket struct {
	client *client.Client
}

// Create a Bucket.
func (h Bucket) Create(r *api.Bucket) (err error) {
	err = h.client.Post(api.BucketsRoute, r)
	return
}

// Get a bucket.
func (h Bucket) Get(id uint) (r *api.Bucket, err error) {
	r = &api.Bucket{}
	path := client.Path(api.BucketRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List buckets.
func (h Bucket) List() (list []api.Bucket, err error) {
	list = []api.Bucket{}
	err = h.client.Get(api.BucketsRoute, &list)
	return
}

// Delete a bucket.
func (h Bucket) Delete(id uint) (err error) {
	path := client.Path(api.BucketRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// Content returns content API.
// Deprecated.  Use Selected().
func (h Bucket) Content(id uint) (b Content) {
	selected := h.Select(id)
	b = selected.Content
	return
}

// Select returns the API for the selected bucket.
func (h Bucket) Select(id uint) (h2 *Selected) {
	h2 = &Selected{}
	root := client.Path(api.BucketRoute).
		Inject(client.Params{
			api.Wildcard: "",
			api.ID:       id,
		})
	h2.Content = NewContent(h.client, root)
	return
}

// Selected bucket API.
type Selected struct {
	Content Content
}

// Content API.
type Content struct {
	client *client.Client
	root   string
}

// Get reads from the bucket.
// The source (root) is relative to the bucket root.
func (h Content) Get(source, destination string) (err error) {
	err = h.client.BucketGet(pathlib.Join(h.root, source), destination)
	return
}

// Put writes to the bucket.
// The destination (root) is relative to the bucket root.
func (h Content) Put(source, destination string) (err error) {
	err = h.client.BucketPut(source, pathlib.Join(h.root, destination))
	return
}

// Delete deletes content at the specified root.
// The source is relative to the bucket root.
func (h Content) Delete(path string) (err error) {
	err = h.client.Delete(pathlib.Join(h.root, path))
	return
}
