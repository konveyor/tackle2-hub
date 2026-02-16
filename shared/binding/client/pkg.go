package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/konveyor/tackle2-hub/shared/api"
	qf "github.com/konveyor/tackle2-hub/shared/binding/filter"
)

const (
	RetryLimit = 60
	RetryDelay = time.Second * 10
)

// New Constructs a new client
func New(baseURL string) (client *Client) {
	client = &Client{
		BaseURL: baseURL,
	}
	client.Retry = RetryLimit
	return
}

// Param http parameter.
type Param struct {
	Key   string
	Value string
}

// Filter filter list results.
type Filter struct {
	qf.Filter
}

// Param returns a filter parameter.
func (r *Filter) Param() (p Param) {
	p.Key = "filter"
	p.Value = r.String()
	return
}

// Params mapping.
type Params map[string]any

// Path API path.
type Path string

// Inject named parameters.
func (s Path) Inject(p Params) (out string) {
	in := strings.Split(string(s), "/")
	for i := range in {
		if len(in[i]) < 1 {
			continue
		}
		key := in[i][1:]
		if v, found := p[key]; found {
			in[i] = fmt.Sprintf("%v", v)
		}
	}
	out = strings.Join(in, "/")
	return
}

// RestClient defines the interface for REST client operations.
// It provides methods for standard HTTP operations (GET, POST, PUT, PATCH, DELETE),
// file/bucket operations, and utility functions.
type RestClient interface {
	// Reset clears the error state of the client.
	Reset()
	// Use login.
	Use(login api.Login)
	// SetRetry set the number of retries.
	SetRetry(n uint8)

	// Get retrieves a resource from the specified path.
	// The response is unmarshaled into the provided object.
	// Optional query parameters can be provided via params.
	Get(path string, object any, params ...Param) (err error)

	// Post creates a resource at the specified path.
	// The object is marshaled and sent as the request body.
	// The response is unmarshaled back into the object.
	Post(path string, object any) (err error)

	// Put updates a resource at the specified path.
	// The object is marshaled and sent as the request body.
	// The response is unmarshaled back into the object.
	// Optional query parameters can be provided via params.
	Put(path string, object any, params ...Param) (err error)

	// Patch partially updates a resource at the specified path.
	// The object is marshaled and sent as the request body.
	// The response is unmarshaled back into the object.
	// Optional query parameters can be provided via params.
	Patch(path string, object any, params ...Param) (err error)

	// Delete removes a resource at the specified path.
	// Optional query parameters can be provided via params.
	Delete(path string, params ...Param) (err error)

	// DeleteWith removes a resource at the specified path as specified by the body.
	// Optional query parameters can be provided via params.
	DeleteWith(path string, body any, params ...Param) (err error)

	// BucketGet downloads a file or directory from the bucket.
	// The source path is relative to the bucket root.
	// Directories are automatically extracted to the destination.
	BucketGet(source, destination string) (err error)

	// BucketPut uploads a file or directory to the bucket.
	// The destination path is relative to the bucket root.
	// Directories are automatically archived before upload.
	BucketPut(source, destination string) (err error)

	// FileGet downloads a file from the specified path.
	FileGet(path, destination string) (err error)

	// FilePost uploads a file to the specified path using POST.
	// Returns the created File resource in object.
	FilePost(path, source string, object any) (err error)

	// FilePostEncoded uploads a file with a specific encoding using POST.
	FilePostEncoded(path, source string, object any, encoding string) (err error)

	// FilePut uploads a file to the specified path using PUT.
	FilePut(path, source string, object any) (err error)

	// FilePutEncoded uploads a file with a specific encoding using PUT.
	FilePutEncoded(path, source string, object any, encoding string) (err error)

	// FilePatch appends data to a file at the specified path.
	FilePatch(path string, buffer []byte) (err error)

	// FileSend sends a multipart file upload request.
	// The method parameter specifies the HTTP method to use.
	// Fields contains the form fields to include in the upload.
	// The response is unmarshaled into object if provided.
	FileSend(path, method string, fields []Field, object any) (err error)

	// IsDir determines if the given path is a directory.
	// If must is true, the path must exist or an error is returned.
	// If must is false, non-existent paths return false without error.
	IsDir(path string, must bool) (isDir bool, err error)
}
