package binding

import (
	"github.com/konveyor/tackle2-hub/api"
	pathlib "path"
)

//
// File API.
type File struct {
	client *Client
}

//
// Get downloads a file.
func (h *File) Get(id uint, destination string) (err error) {
	path := Path(api.FileRoot).Inject(Params{api.ID: id})
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

//
// Put uploads a file.
func (h *File) Put(source string) (r *api.File, err error) {
	r = &api.File{}
	path := Path(api.FileRoot).Inject(Params{api.ID: pathlib.Base(source)})
	err = h.client.FilePut(path, source, r)
	return
}

//
// Delete a file.
func (h *File) Delete(id uint) (err error) {
	path := Path(api.FileRoot).Inject(Params{api.ID: id})
	err = h.client.Delete(path)
	return
}
