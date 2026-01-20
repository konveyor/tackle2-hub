package client

var _ RestClient = (*Stub)(nil)

// Stub base implementation intended to support overrides for testing.
type Stub struct{}

// Reset clears the error state of the client.
func (c *Stub) Reset() {
	// no-op
}

// Get retrieves a resource from the specified path.
func (c *Stub) Get(path string, object any, params ...Param) (err error) {
	return
}

// Post creates a resource at the specified path.
func (c *Stub) Post(path string, object any) (err error) {
	return
}

// Put updates a resource at the specified path.
func (c *Stub) Put(path string, object any, params ...Param) (err error) {
	return
}

// Patch partially updates a resource at the specified path.
func (c *Stub) Patch(path string, object any, params ...Param) (err error) {
	return
}

// Delete removes a resource at the specified path.
func (c *Stub) Delete(path string, params ...Param) (err error) {
	return
}

// BucketGet downloads a file or directory from the bucket.
func (c *Stub) BucketGet(source, destination string) (err error) {
	return
}

// BucketPut uploads a file or directory to the bucket.
func (c *Stub) BucketPut(source, destination string) (err error) {
	return
}

// FileGet downloads a file from the specified path.
func (c *Stub) FileGet(path, destination string) (err error) {
	return
}

// FilePost uploads a file to the specified path using POST.
func (c *Stub) FilePost(path, source string, object any) (err error) {
	return
}

// FilePostEncoded uploads a file with a specific encoding using POST.
func (c *Stub) FilePostEncoded(path, source string, object any, encoding string) (err error) {
	return
}

// FilePut uploads a file to the specified path using PUT.
func (c *Stub) FilePut(path, source string, object any) (err error) {
	return
}

// FilePutEncoded uploads a file with a specific encoding using PUT.
func (c *Stub) FilePutEncoded(path, source string, object any, encoding string) (err error) {
	return
}

// FilePatch appends data to a file at the specified path.
func (c *Stub) FilePatch(path string, buffer []byte) (err error) {
	return
}

// FileSend sends a multipart file upload request.
func (c *Stub) FileSend(path, method string, fields []Field, object any) (err error) {
	return
}

// IsDir determines if the given path is a directory.
func (c *Stub) IsDir(path string, must bool) (isDir bool, err error) {
	return false, nil
}
