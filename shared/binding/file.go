package binding

import (
	pathlib "path"

	"github.com/konveyor/tackle2-hub/shared/api"
)

// File API.
type File struct {
	client *Client
}

// Get downloads a file.
func (h File) Get(id uint, destination string) (err error) {
	path := Path(api.FileRoute).Inject(Params{api.ID: id})
	isDir, err := h.client.IsDir(destination, false)
	if err != nil {
		return
	}
	if isDir {
		r := &api.File{}
		err = h.client.Get(path, r)
		if err != nil {
			return
		}
		destination = pathlib.Join(
			destination,
			r.Name)
	}
	err = h.client.FileGet(path, destination)
	return
}

// Touch creates an empty file.
func (h File) Touch(name string) (r *api.File, err error) {
	r = &api.File{}
	path := Path(api.FileRoute).Inject(Params{api.ID: name})
	err = h.client.FilePost(path, "", r)
	return
}

// Post uploads a file.
func (h File) Post(source string) (r *api.File, err error) {
	r, err = h.PostEncoded(source, "")
	return
}

// PostEncoded uploads a file.
func (h File) PostEncoded(source string, encoding string) (r *api.File, err error) {
	r = &api.File{}
	path := Path(api.FileRoute).Inject(Params{api.ID: pathlib.Base(source)})
	err = h.client.FilePostEncoded(path, source, r, encoding)
	return
}

// Put uploads a file.
func (h File) Put(source string) (r *api.File, err error) {
	r, err = h.PutEncoded(source, "")
	return
}

// PutEncoded uploads a file.
func (h File) PutEncoded(source string, encoding string) (r *api.File, err error) {
	r = &api.File{}
	path := Path(api.FileRoute).Inject(Params{api.ID: pathlib.Base(source)})
	err = h.client.FilePutEncoded(path, source, r, encoding)
	return
}

// Patch appends a file.
func (h File) Patch(id uint, buffer []byte) (err error) {
	path := Path(api.FileRoute).Inject(Params{api.ID: id})
	err = h.client.FilePatch(path, buffer)
	return
}

// Delete a file.
func (h File) Delete(id uint) (err error) {
	path := Path(api.FileRoute).Inject(Params{api.ID: id})
	err = h.client.Delete(path)
	return
}
