package addon

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin/binding"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/api"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
	"time"
)

const (
	RetryLimit = 60
	RetryDelay = time.Second * 10
)

//
// Param.
type Param struct {
	Key   string
	Value string
}

//
// Params mapping.
type Params map[string]interface{}

//
// Path API path.
type Path string

//
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

//
// NewClient Constructs a new client
func NewClient(url, token string) (client *Client) {
	client = &Client{
		baseURL: url,
		token:   token,
	}
	client.Retry = RetryLimit
	return
}

//
// Client provides a REST client.
type Client struct {
	// baseURL for the nub.
	baseURL string
	// addon API token
	token string
	// transport
	transport http.RoundTripper
	// Retry limit.
	Retry int
	// Error
	Error error
}

//
// SetToken sets hub token on client
func (r *Client) SetToken(token string) {
	r.token = token
}

//
// Reset the client.
func (r *Client) Reset() {
	r.Error = nil
}

//
// Get a resource.
func (r *Client) Get(path string, object interface{}, params ...Param) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodGet,
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, binding.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = reply.Body.Close()
	}()
	status := reply.StatusCode
	switch status {
	case http.StatusOK:
		var body []byte
		body, err = io.ReadAll(reply.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case http.StatusNotFound:
		err = &NotFound{Path: path}
	default:
		err = liberr.New(http.StatusText(status))
	}

	return
}

//
// Post a resource.
func (r *Client) Post(path string, object interface{}) (err error) {
	request := func() (request *http.Request, err error) {
		bfr, err := json.Marshal(object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		reader := bytes.NewReader(bfr)
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodPost,
			Body:   io.NopCloser(reader),
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, binding.MIMEJSON)
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	status := reply.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(reply.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case http.StatusConflict:
		err = &Conflict{Path: path}
	default:
		err = liberr.New(http.StatusText(status))
	}

	return
}

//
// Put a resource.
func (r *Client) Put(path string, object interface{}, params ...Param) (err error) {
	request := func() (request *http.Request, err error) {
		bfr, err := json.Marshal(object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		reader := bytes.NewReader(bfr)
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodPut,
			Body:   io.NopCloser(reader),
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, binding.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	status := reply.StatusCode
	switch status {
	case http.StatusNoContent:
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(reply.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case http.StatusNotFound:
		err = &NotFound{Path: path}
	default:
		err = liberr.New(http.StatusText(status))
	}

	return
}

//
// Delete a resource.
func (r *Client) Delete(path string, params ...Param) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodDelete,
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, binding.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = reply.Body.Close()
	}()
	status := reply.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusNoContent:
	case http.StatusNotFound:
		err = &NotFound{Path: path}
	default:
		err = liberr.New(http.StatusText(status))
	}

	return
}

//
// BucketGet downloads a file/directory.
// The source (path) is relative to the bucket root.
func (r *Client) BucketGet(source, destination string) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodGet,
			URL:    r.join(source),
		}
		request.Header.Set(api.Accept, api.MIMEOCTETSTREAM)
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = reply.Body.Close()
	}()
	status := reply.StatusCode
	switch status {
	case http.StatusNoContent:
		// Empty.
	case http.StatusOK:
		if reply.Header.Get(api.Directory) == api.DirectoryExpand {
			err = r.getDir(reply.Body, destination)
		} else {
			err = r.getFile(reply.Body, source, destination)
		}
	case http.StatusNotFound:
		err = &NotFound{Path: source}
	default:
		err = liberr.New(http.StatusText(status))
	}
	return
}

//
// BucketPut uploads a file/directory.
// The destination (path) is relative to the bucket root.
func (r *Client) BucketPut(source, destination string) (err error) {
	isDir, err := r.isDir(source, true)
	if err != nil {
		return
	}
	request := func() (request *http.Request, err error) {
		buf := new(bytes.Buffer)
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodPut,
			Body:   io.NopCloser(buf),
			URL:    r.join(destination),
		}
		request.Header.Set(api.Accept, api.MIMEOCTETSTREAM)
		writer := multipart.NewWriter(buf)
		defer func() {
			_ = writer.Close()
		}()
		part, nErr := writer.CreateFormFile(api.FileField, pathlib.Base(source))
		if err != nil {
			err = liberr.Wrap(nErr)
			return
		}
		request.Header.Add(
			api.ContentType,
			writer.FormDataContentType())
		if isDir {
			request.Header.Set(api.Directory, api.DirectoryExpand)
			err = r.putDir(part, source)
		} else {
			err = r.putFile(part, source)
		}
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	status := reply.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusNoContent,
		http.StatusCreated,
		http.StatusAccepted:
	case http.StatusNotFound:
		err = &NotFound{Path: destination}
	default:
		err = liberr.New(http.StatusText(status))
	}
	return
}

//
// FileGet downloads a file.
func (r *Client) FileGet(path, destination string) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodGet,
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, api.MIMEOCTETSTREAM)
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = reply.Body.Close()
	}()
	status := reply.StatusCode
	switch status {
	case http.StatusNoContent:
		// Empty.
	case http.StatusOK:
		err = r.getFile(reply.Body, "", destination)
	case http.StatusNotFound:
		err = &NotFound{Path: path}
	default:
		err = liberr.New(http.StatusText(status))
	}
	return
}

//
// FilePut uploads a file.
// Returns the created File resource.
func (r *Client) FilePut(path, source string, object interface{}) (err error) {
	isDir, err := r.isDir(source, true)
	if err != nil {
		return
	}
	if isDir {
		err = liberr.New("Source cannot be directory.")
		return
	}
	request := func() (request *http.Request, err error) {
		buf := new(bytes.Buffer)
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodPut,
			Body:   io.NopCloser(buf),
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, binding.MIMEJSON)
		writer := multipart.NewWriter(buf)
		defer func() {
			_ = writer.Close()
		}()
		part, nErr := writer.CreateFormFile(api.FileField, pathlib.Base(source))
		if err != nil {
			err = liberr.Wrap(nErr)
			return
		}
		request.Header.Add(
			api.ContentType,
			writer.FormDataContentType())
		err = r.putFile(part, source)
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	status := reply.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(reply.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	case http.StatusConflict:
		err = &Conflict{Path: path}
	default:
		err = liberr.New(http.StatusText(status))
	}
	return
}

//
// getDir downloads and expands a directory.
func (r *Client) getDir(body io.Reader, output string) (err error) {
	zipReader, err := gzip.NewReader(body)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = zipReader.Close()
	}()
	tarReader := tar.NewReader(zipReader)
	for {
		header, nErr := tarReader.Next()
		if nErr != nil {
			if nErr == io.EOF {
				break
			} else {
				err = liberr.Wrap(nErr)
				return
			}
		}
		path := pathlib.Join(output, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.Mkdir(path, 0777)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		case tar.TypeReg:
			file, nErr := os.Create(path)
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			_, err = io.Copy(file, tarReader)
			_ = file.Close()
		default:
		}
	}
	return
}

//
// putDir archive and uploads a directory.
func (r *Client) putDir(writer io.Writer, input string) (err error) {
	var tarOutput bytes.Buffer
	tarWriter := tar.NewWriter(&tarOutput)
	err = filepath.Walk(
		input,
		func(path string, entry os.FileInfo, wErr error) (err error) {
			if wErr != nil {
				err = liberr.Wrap(wErr)
				return
			}
			if path == input {
				return
			}
			header, nErr := tar.FileInfoHeader(entry, "")
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			header.Name = path[len(input)+1:]
			switch header.Typeflag {
			case tar.TypeDir:
				err = tarWriter.WriteHeader(header)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
			case tar.TypeReg:
				err = tarWriter.WriteHeader(header)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				file, nErr := os.Open(path)
				if err != nil {
					err = liberr.Wrap(nErr)
					return
				}
				defer func() {
					_ = file.Close()
				}()
				_, err = io.Copy(tarWriter, file)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
			}
			return
		})
	if err != nil {
		return
	}
	zipReader := bufio.NewReader(&tarOutput)
	zipWriter := gzip.NewWriter(writer)
	defer func() {
		_ = zipWriter.Close()
	}()
	_, err = io.Copy(zipWriter, zipReader)
	return
}

//
// getFile downloads plain file.
func (r *Client) getFile(body io.Reader, path, output string) (err error) {
	isDir, err := r.isDir(output, false)
	if err != nil {
		return
	}
	if isDir {
		output = pathlib.Join(
			output,
			pathlib.Base(path))
	}
	file, err := os.Create(output)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = io.Copy(file, body)
	return
}

//
// putFile uploads plain file.
func (r *Client) putFile(writer io.Writer, input string) (err error) {
	file, err := os.Open(input)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = io.Copy(writer, file)
	return
}

//
// isDir determines if the path is a directory.
// The `must` specifies if the path must exist.
func (r *Client) isDir(path string, must bool) (b bool, err error) {
	st, err := os.Stat(path)
	if err == nil {
		b = st.IsDir()
		return
	}
	if os.IsNotExist(err) {
		if must {
			err = liberr.Wrap(err)
		} else {
			err = nil
		}
	} else {
		err = liberr.Wrap(err)
	}
	return
}

//
// Send the request.
// Resilient against transient hub availability.
func (r *Client) send(rb func() (*http.Request, error)) (response *http.Response, err error) {
	var request *http.Request
	if r.Error != nil {
		err = r.Error
		return
	}
	err = r.buildTransport()
	if err != nil {
		return
	}
	for i := 0; ; i++ {
		request, err = rb()
		if err != nil {
			return
		}
		request.Header.Set(api.Authorization, r.token)
		client := http.Client{Transport: r.transport}
		response, err = client.Do(request)
		if err != nil {
			netErr := &net.OpError{}
			if errors.As(err, &netErr) {
				if i < r.Retry {
					Log.Info(err.Error())
					time.Sleep(RetryDelay)
					continue
				} else {
					r.Error = liberr.Wrap(err)
					err = r.Error
					return
				}
			} else {
				err = liberr.Wrap(err)
				return
			}
		} else {
			Log.Info(
				fmt.Sprintf(
					"|%d|  %s %s",
					response.StatusCode,
					request.Method,
					request.URL.Path))
			break
		}
	}
	return
}

//
// buildTransport builds transport.
func (r *Client) buildTransport() (err error) {
	if r.transport != nil {
		return
	}
	r.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		MaxIdleConns:          3,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return
}

//
// Join the URL.
func (r *Client) join(path string) (parsedURL *url.URL) {
	parsedURL, _ = url.Parse(r.baseURL)
	parsedURL.Path = pathlib.Join(parsedURL.Path, path)
	return
}
