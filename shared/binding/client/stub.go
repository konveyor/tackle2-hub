package client

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

var _ RestClient = (*Stub)(nil)

// Stub implementation
type Stub struct {
	GetFn    func(path string, object any, params ...Param) (err error)
	PostFn   func(path string, object any) (err error)
	PutFn    func(path string, object any, params ...Param) (err error)
	PatchFn  func(path string, object any, params ...Param) (err error)
	DeleteFn func(path string, params ...Param) (err error)

	BucketGetFn func(source, destination string) (err error)
	BucketPutFn func(source, destination string) (err error)

	FileGetFn         func(path, destination string) (err error)
	FilePostFn        func(path, source string, object any) (err error)
	FilePostEncodedFn func(path, source string, object any, encoding string) (err error)
	FilePutFn         func(path, source string, object any) (err error)
	FilePutEncodedFn  func(path, source string, object any, encoding string) (err error)
	FilePatchFn       func(path string, buffer []byte) (err error)
	FileSendFn        func(path, method string, fields []Field, object any) (err error)

	IsDirFn func(path string, must bool) (isDir bool, err error)
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
	if s.GetFn == nil {
		panic("Get not implemented")
	}
	return s.GetFn(path, object, params...)
}

// Post creates a resource at the specified path.
func (s *Stub) Post(path string, object any) (err error) {
	if s.PostFn == nil {
		panic("Post not implemented")
	}
	return s.PostFn(path, object)
}

// Put updates a resource at the specified path.
func (s *Stub) Put(path string, object any, params ...Param) (err error) {
	if s.PutFn == nil {
		panic("Put not implemented")
	}
	return s.PutFn(path, object, params...)
}

// Patch partially updates a resource at the specified path.
func (s *Stub) Patch(path string, object any, params ...Param) (err error) {
	if s.PatchFn == nil {
		panic("Patch not implemented")
	}
	return s.PatchFn(path, object, params...)
}

// Delete removes a resource at the specified path.
func (s *Stub) Delete(path string, params ...Param) (err error) {
	if s.DeleteFn == nil {
		panic("Delete not implemented")
	}
	return s.DeleteFn(path, params...)
}

// BucketGet downloads a file or directory from the bucket.
func (s *Stub) BucketGet(source, destination string) (err error) {
	if s.BucketGetFn == nil {
		panic("BucketGet not implemented")
	}
	return s.BucketGetFn(source, destination)
}

// BucketPut uploads a file or directory to the bucket.
func (s *Stub) BucketPut(source, destination string) (err error) {
	if s.BucketPutFn == nil {
		panic("BucketPut not implemented")
	}
	return s.BucketPutFn(source, destination)
}

// FileGet downloads a file from the specified path.
func (s *Stub) FileGet(path, destination string) (err error) {
	if s.FileGetFn == nil {
		panic("FileGet not implemented")
	}
	return s.FileGetFn(path, destination)
}

// FilePost uploads a file to the specified path using POST.
func (s *Stub) FilePost(path, source string, object any) (err error) {
	if s.FilePostFn == nil {
		panic("FilePost not implemented")
	}
	return s.FilePostFn(path, source, object)
}

// FilePostEncoded uploads a file with a specific encoding using POST.
func (s *Stub) FilePostEncoded(path, source string, object any, encoding string) (err error) {
	if s.FilePostEncodedFn == nil {
		panic("FilePostEncoded not implemented")
	}
	return s.FilePostEncodedFn(path, source, object, encoding)
}

// FilePut uploads a file to the specified path using PUT.
func (s *Stub) FilePut(path, source string, object any) (err error) {
	if s.FilePutFn == nil {
		panic("FilePut not implemented")
	}
	return s.FilePutFn(path, source, object)
}

// FilePutEncoded uploads a file with a specific encoding using PUT.
func (s *Stub) FilePutEncoded(path, source string, object any, encoding string) (err error) {
	if s.FilePutEncodedFn == nil {
		panic("FilePutEncoded not implemented")
	}
	return s.FilePutEncodedFn(path, source, object, encoding)
}

// FilePatch appends data to a file at the specified path.
func (s *Stub) FilePatch(path string, buffer []byte) (err error) {
	if s.FilePatchFn == nil {
		panic("FilePatch not implemented")
	}
	return s.FilePatchFn(path, buffer)
}

// FileSend sends a multipart file upload request.
func (s *Stub) FileSend(path, method string, fields []Field, object any) (err error) {
	if s.FileSendFn == nil {
		panic("FileSend not implemented")
	}
	return s.FileSendFn(path, method, fields, object)
}

// IsDir determines if the given path is a directory.
func (s *Stub) IsDir(path string, must bool) (isDir bool, err error) {
	if s.IsDirFn == nil {
		panic("IsDir not implemented")
	}
	return s.IsDirFn(path, must)
}
