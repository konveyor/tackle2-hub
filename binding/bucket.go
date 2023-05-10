package binding

import pathlib "path"

//
// Bucket API.
type Bucket struct {
	// hub API client.
	client *Client
	// root path
	path string
}

//
// Get reads from the bucket.
// The source (path) is relative to the bucket root.
func (h *Bucket) Get(source, destination string) (err error) {
	err = h.client.BucketGet(pathlib.Join(h.path, source), destination)
	return
}

//
// Put writes to the bucket.
// The destination (path) is relative to the bucket root.
func (h *Bucket) Put(source, destination string) (err error) {
	err = h.client.BucketPut(source, pathlib.Join(h.path, destination))
	return
}

//
// Delete deletes content at the specified path.
// The path is relative to the bucket root.
func (h *Bucket) Delete(path string) (err error) {
	err = h.client.Delete(pathlib.Join(h.path, path))
	return
}
