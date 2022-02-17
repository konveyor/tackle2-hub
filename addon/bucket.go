package addon

import (
	"errors"
	"github.com/konveyor/tackle2-hub/api"
	"os"
	pathlib "path"
)

//
// Bucket API.
type Bucket struct {
	// hub API client.
	client *Client
}

//
// Create a bucket.
func (h *Bucket) Create(r *api.Bucket) (err error) {
	err = h.client.Post(api.BucketsRoot, r)
	if err == nil {
		Log.Info(
			"Addon created: bucket.",
			"object",
			r)
	}
	return
}

//
// Ensure a bucket by application and name.
func (h *Bucket) Ensure(appId uint, name string) (r *api.Bucket, err error) {
	r = &api.Bucket{}
	params := Params{
		api.ID:   appId,
		api.Name: name,
	}
	path := params.inject(api.AppBucketRoot)
	err = h.client.Post(path, r)
	if errors.Is(err, &Conflict{}) {
		err = h.client.Get(path, r)
	}
	if err == nil {
		Log.Info(
			"Addon ensured: bucket.",
			"object",
			r)
	}
	return
}

//
// Get a bucket by ID.
func (h *Bucket) Get(id uint) (r *api.Bucket, err error) {
	r = &api.Bucket{}
	path := Params{api.ID: id}.inject(api.BucketsRoot)
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
// Delete an bucket.
func (h *Bucket) Delete(r *api.Bucket) (err error) {
	path := Params{api.ID: r.ID}.inject(api.BucketsRoot)
	err = h.client.Delete(path)
	if err == nil {
		Log.Info(
			"Addon deleted: bucket.",
			"object",
			r)
	}
	return
}

//
// Purge bucket.
func (h *Bucket) Purge(r *api.Bucket) (err error) {
	dir, err := os.ReadDir(r.Path)
	if err != nil {
		return
	}
	for _, p := range dir {
		path := pathlib.Join(r.Path, p.Name())
		err = os.RemoveAll(path)
		if err != nil {
			return
		}
	}
	return
}
