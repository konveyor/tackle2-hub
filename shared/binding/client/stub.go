package client

import (
	"fmt"

	"github.com/konveyor/tackle2-hub/shared/api"
)

var _ RestClient = (*Stub)(nil)

// Stub implementation
type Stub struct {
	DoGet        func(path string, object any, params ...Param) (err error)
	DoPost       func(path string, object any) (err error)
	DoPut        func(path string, object any, params ...Param) (err error)
	DoPatch      func(path string, object any, params ...Param) (err error)
	DoDeleteWith func(path string, body any, params ...Param) (err error)
	DoDelete     func(path string, params ...Param) (err error)

	DoBucketGet func(source, destination string) (err error)
	DoBucketPut func(source, destination string) (err error)

	DoFileGet         func(path, destination string) (err error)
	DoFilePost        func(path, source string, object any) (err error)
	DoFilePostEncoded func(path, source string, object any, encoding string) (err error)
	DoFilePut         func(path, source string, object any) (err error)
	DoFilePutEncoded  func(path, source string, object any, encoding string) (err error)
	DoFilePatch       func(path string, buffer []byte) (err error)
	DoFileSend        func(path, method string, fields []Field, object any) (err error)

	DoIsDir func(path string, must bool) (isDir bool, err error)
}

// Reset clears the error state of the client.
func (s *Stub) Reset() {
}

// Use login.
func (s *Stub) Use(login api.Login) {
}

// SetRetry set the number of retries.
func (s *Stub) SetRetry(n uint8) {
}

// Get retrieves a resource from the specified path.
func (s *Stub) Get(path string, object any, params ...Param) (err error) {
	if s.DoGet == nil {
		err = fmt.Errorf("Get not implemented")
		return
	}
	return s.DoGet(path, object, params...)
}

// Post creates a resource at the specified path.
func (s *Stub) Post(path string, object any) (err error) {
	if s.DoPost == nil {
		err = fmt.Errorf("Post not implemented")
		return
	}
	return s.DoPost(path, object)
}

// Put updates a resource at the specified path.
func (s *Stub) Put(path string, object any, params ...Param) (err error) {
	if s.DoPut == nil {
		err = fmt.Errorf("Put not implemented")
		return
	}
	return s.DoPut(path, object, params...)
}

// Patch partially updates a resource at the specified path.
func (s *Stub) Patch(path string, object any, params ...Param) (err error) {
	if s.DoPatch == nil {
		err = fmt.Errorf("Patch not implemented")
		return
	}
	return s.DoPatch(path, object, params...)
}

// Delete removes a resource at the specified path.
func (s *Stub) Delete(path string, params ...Param) (err error) {
	if s.DoDelete == nil {
		err = fmt.Errorf("Delete not implemented")
		return
	}
	return s.DoDelete(path, params...)
}

// DeleteWith removes a resource at the specified path as spcified by the body.
func (s *Stub) DeleteWith(path string, body any, params ...Param) (err error) {
	if s.DoDeleteWith == nil {
		err = fmt.Errorf("DeleteWith not implemented")
		return
	}
	return s.DoDeleteWith(path, body, params...)
}

// BucketGet downloads a file or directory from the bucket.
func (s *Stub) BucketGet(source, destination string) (err error) {
	if s.DoBucketGet == nil {
		err = fmt.Errorf("BucketGet not implemented")
		return
	}
	return s.DoBucketGet(source, destination)
}

// BucketPut uploads a file or directory to the bucket.
func (s *Stub) BucketPut(source, destination string) (err error) {
	if s.DoBucketPut == nil {
		err = fmt.Errorf("BucketPut not implemented")
		return
	}
	return s.DoBucketPut(source, destination)
}

// FileGet downloads a file from the specified path.
func (s *Stub) FileGet(path, destination string) (err error) {
	if s.DoFileGet == nil {
		err = fmt.Errorf("FileGet not implemented")
		return
	}
	return s.DoFileGet(path, destination)
}

// FilePost uploads a file to the specified path using POST.
func (s *Stub) FilePost(path, source string, object any) (err error) {
	if s.DoFilePost == nil {
		err = fmt.Errorf("FilePost not implemented")
		return
	}
	return s.DoFilePost(path, source, object)
}

// FilePostEncoded uploads a file with a specific encoding using POST.
func (s *Stub) FilePostEncoded(path, source string, object any, encoding string) (err error) {
	if s.DoFilePostEncoded == nil {
		err = fmt.Errorf("FilePostEncoded not implemented")
		return
	}
	return s.DoFilePostEncoded(path, source, object, encoding)
}

// FilePut uploads a file to the specified path using PUT.
func (s *Stub) FilePut(path, source string, object any) (err error) {
	if s.DoFilePut == nil {
		err = fmt.Errorf("FilePut not implemented")
		return
	}
	return s.DoFilePut(path, source, object)
}

// FilePutEncoded uploads a file with a specific encoding using PUT.
func (s *Stub) FilePutEncoded(path, source string, object any, encoding string) (err error) {
	if s.DoFilePutEncoded == nil {
		err = fmt.Errorf("FilePutEncoded not implemented")
		return
	}
	return s.DoFilePutEncoded(path, source, object, encoding)
}

// FilePatch appends data to a file at the specified path.
func (s *Stub) FilePatch(path string, buffer []byte) (err error) {
	if s.DoFilePatch == nil {
		err = fmt.Errorf("FilePatch not implemented")
		return
	}
	return s.DoFilePatch(path, buffer)
}

// FileSend sends a multipart file upload request.
func (s *Stub) FileSend(path, method string, fields []Field, object any) (err error) {
	if s.DoFileSend == nil {
		err = fmt.Errorf("FileSend not implemented")
		return
	}
	return s.DoFileSend(path, method, fields, object)
}

// IsDir determines if the given path is a directory.
func (s *Stub) IsDir(path string, must bool) (isDir bool, err error) {
	if s.DoIsDir == nil {
		isDir = false
		err = fmt.Errorf("IsDir not implemented")
		return
	}
	return s.DoIsDir(path, must)
}
