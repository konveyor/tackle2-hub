package binding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	pathlib "path"
	"strings"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/api"
	qf "github.com/konveyor/tackle2-hub/shared/binding/filter"
	"github.com/konveyor/tackle2-hub/shared/tar"
)

const (
	RetryLimit = 60
	RetryDelay = time.Second * 10
)

// Param.
type Param struct {
	Key   string
	Value string
}

// Filter
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

// NewClient Constructs a new client
func NewClient(baseURL string) (client *Client) {
	client = &Client{
		BaseURL: baseURL,
	}
	client.Retry = RetryLimit
	return
}

// Client provides a REST client.
type Client struct {
	// transport
	transport http.RoundTripper
	// baseURL for the nub.
	BaseURL string
	// login API resource.
	Login api.Login
	// Retry limit.
	Retry int
	// Error
	Error error
}

// Reset the client.
func (r *Client) Reset() {
	r.Error = nil
}

// Get a resource.
func (r *Client) Get(path string, object any, params ...Param) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodGet,
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, api.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()
	status := response.StatusCode
	switch status {
	case http.StatusOK:
		var body []byte
		body, err = io.ReadAll(response.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if len(body) == 0 {
			empty := &EmptyBody{}
			empty.With(response)
			err = empty
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		err = r.restError(response)
	}

	return
}

// Post a resource.
func (r *Client) Post(path string, object any) (err error) {
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
		request.Header.Set(api.Accept, api.MIMEJSON)
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	status := response.StatusCode
	switch status {
	case http.StatusAccepted:
	case http.StatusNoContent:
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(response.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if len(body) == 0 {
			empty := &EmptyBody{}
			empty.With(response)
			err = empty
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		err = r.restError(response)
	}
	return
}

// Put a resource.
func (r *Client) Put(path string, object any, params ...Param) (err error) {
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
		request.Header.Set(api.Accept, api.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	status := response.StatusCode
	switch status {
	case http.StatusAccepted:
	case http.StatusNoContent:
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(response.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if len(body) == 0 {
			empty := &EmptyBody{}
			empty.With(response)
			err = empty
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		err = r.restError(response)
	}

	return
}

// Patch a resource.
func (r *Client) Patch(path string, object any, params ...Param) (err error) {
	request := func() (request *http.Request, err error) {
		bfr, err := json.Marshal(object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		reader := bytes.NewReader(bfr)
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodPatch,
			Body:   io.NopCloser(reader),
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, api.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	status := response.StatusCode
	switch status {
	case http.StatusAccepted:
	case http.StatusNoContent:
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(response.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if len(body) == 0 {
			empty := &EmptyBody{}
			empty.With(response)
			err = empty
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		err = r.restError(response)
	}

	return
}

// Delete a resource.
func (r *Client) Delete(path string, params ...Param) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodDelete,
			URL:    r.join(path),
		}
		request.Header.Set(api.Accept, api.MIMEJSON)
		if len(params) > 0 {
			q := request.URL.Query()
			for _, p := range params {
				q.Add(p.Key, p.Value)
			}
			request.URL.RawQuery = q.Encode()
		}
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()
	status := response.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusNoContent:
	default:
		err = r.restError(response)
	}

	return
}

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
	response, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()
	status := response.StatusCode
	switch status {
	case http.StatusNoContent:
		// Empty.
	case http.StatusOK:
		if response.Header.Get(api.Directory) == api.DirectoryExpand {
			err = r.getDir(response.Body, destination)
		} else {
			err = r.getFile(response.Body, source, destination)
		}
	default:
		err = r.restError(response)
	}
	return
}

// BucketPut uploads a file/directory.
// The destination (path) is relative to the bucket root.
func (r *Client) BucketPut(source, destination string) (err error) {
	isDir, err := r.IsDir(source, true)
	if err != nil {
		return
	}
	request := func() (request *http.Request, err error) {
		pr, pw := io.Pipe()
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodPut,
			Body:   pr,
			URL:    r.join(destination),
		}
		mp := multipart.NewWriter(pw)
		request.Header.Set(api.Accept, api.MIMEOCTETSTREAM)
		request.Header.Add(api.ContentType, mp.FormDataContentType())
		if isDir {
			request.Header.Set(api.Directory, api.DirectoryExpand)
		}
		go func() {
			var err error
			defer func() {
				_ = mp.Close()
				if err != nil {
					_ = pw.CloseWithError(err)
				} else {
					_ = pw.Close()
				}
			}()
			part, nErr := mp.CreateFormFile(api.FileField, pathlib.Base(source))
			if nErr != nil {
				err = nErr
				return
			}
			if isDir {
				err = r.putDir(part, source)
			} else {
				err = r.putFile(part, source)
			}
		}()
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	status := response.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusNoContent,
		http.StatusCreated,
		http.StatusAccepted:
	default:
		err = r.restError(response)
	}
	return
}

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
	response, err := r.send(request)
	if err != nil {
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()
	status := response.StatusCode
	switch status {
	case http.StatusNoContent:
		// Empty.
	case http.StatusOK:
		err = r.getFile(response.Body, "", destination)
	default:
		err = r.restError(response)
	}
	return
}

// FilePost uploads a file.
// Returns the created File resource.
func (r *Client) FilePost(path, source string, object any) (err error) {
	err = r.FilePostEncoded(path, source, object, "")
	return
}

// FilePostEncoded uploads a file.
// Returns the created File resource.
func (r *Client) FilePostEncoded(path, source string, object any, encoding string) (err error) {
	if source == "" {
		fields := []Field{
			{
				Name:     api.FileField,
				Reader:   bytes.NewReader([]byte{}),
				Encoding: encoding,
			},
		}
		err = r.FileSend(path, http.MethodPost, fields, object)
		return
	}
	isDir, nErr := r.IsDir(source, true)
	if nErr != nil {
		err = nErr
		return
	}
	if isDir {
		err = liberr.New("Must be regular file.")
		return
	}
	fields := []Field{
		{
			Name:     api.FileField,
			Path:     source,
			Encoding: encoding,
		},
	}
	err = r.FileSend(path, http.MethodPost, fields, object)
	return
}

// FilePut uploads a file.
// Returns the created File resource.
func (r *Client) FilePut(path, source string, object any) (err error) {
	err = r.FilePutEncoded(path, source, object, "")
	return
}

// FilePutEncoded uploads a file.
// Returns the created File resource.
func (r *Client) FilePutEncoded(path, source string, object any, encoding string) (err error) {
	if source == "" {
		fields := []Field{
			{
				Name:     api.FileField,
				Reader:   bytes.NewReader([]byte{}),
				Encoding: encoding,
			},
		}
		err = r.FileSend(path, http.MethodPut, fields, object)
		return
	}
	isDir, nErr := r.IsDir(source, true)
	if nErr != nil {
		err = nErr
		return
	}
	if isDir {
		err = liberr.New("Must be regular file.")
		return
	}
	fields := []Field{
		{
			Name:     api.FileField,
			Path:     source,
			Encoding: encoding,
		},
	}
	err = r.FileSend(path, http.MethodPut, fields, object)
	return
}

// FilePatch appends file.
// Returns the created File resource.
func (r *Client) FilePatch(path string, buffer []byte) (err error) {
	fields := []Field{
		{
			Name:   api.FileField,
			Reader: bytes.NewReader(buffer),
		},
	}
	err = r.FileSend(path, http.MethodPatch, fields, nil)
	return
}

// FileSend sends file upload from.
func (r *Client) FileSend(path, method string, fields []Field, object any) (err error) {
	request := func() (request *http.Request, err error) {
		pr, pw := io.Pipe()
		request = &http.Request{
			Header: http.Header{},
			Method: method,
			Body:   pr,
			URL:    r.join(path),
		}
		mp := multipart.NewWriter(pw)
		request.Header.Set(api.Accept, api.MIMEJSON)
		request.Header.Add(api.ContentType, mp.FormDataContentType())
		go func() {
			var err error
			defer func() {
				_ = mp.Close()
				if err != nil {
					_ = pw.CloseWithError(err)
				} else {
					_ = pw.Close()
				}
			}()
			for _, f := range fields {
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", f.disposition())
				h.Set("Content-Type", f.encoding())
				part, nErr := mp.CreatePart(h)
				if nErr != nil {
					err = nErr
					return
				}
				err = f.Write(part)
				if err != nil {
					return
				}
			}
		}()
		return
	}
	response, err := r.send(request)
	if err != nil {
		return
	}
	status := response.StatusCode
	switch status {
	case http.StatusOK,
		http.StatusCreated:
		var body []byte
		body, err = io.ReadAll(response.Body)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if len(body) == 0 {
			empty := &EmptyBody{}
			empty.With(response)
			err = empty
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		err = r.restError(response)
	}
	return
}

// getDir downloads and expands a directory.
func (r *Client) getDir(body io.Reader, output string) (err error) {
	tarReader := tar.NewReader()
	err = tarReader.Extract(output, body)
	return
}

// putDir archive and uploads a directory.
func (r *Client) putDir(writer io.Writer, input string) (err error) {
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()
	err = tarWriter.AddDir(input)
	return
}

// getFile downloads plain file.
func (r *Client) getFile(body io.Reader, path, output string) (err error) {
	isDir, err := r.IsDir(output, false)
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

// IsDir determines if the path is a directory.
// The `must` specifies if the path must exist.
func (r *Client) IsDir(path string, must bool) (b bool, err error) {
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
		request.Header.Set(api.Authorization, "Bearer "+r.Login.Token)
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
			if response.StatusCode == http.StatusGatewayTimeout {
				if i < r.Retry {
					_ = response.Body.Close()
					time.Sleep(RetryDelay)
					continue
				}
			}
			if response.StatusCode == http.StatusUnauthorized {
				refreshed, nErr := r.refreshToken(request)
				if nErr != nil {
					r.Error = liberr.Wrap(nErr)
					err = r.Error
					return
				}
				if refreshed {
					_ = response.Body.Close()
					continue
				}
			}
			break
		}
	}
	return
}

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

// Join the URL.
func (r *Client) join(path string) (parsedURL *url.URL) {
	parsedURL, _ = url.Parse(r.BaseURL)
	parsedURL.Path = pathlib.Join(parsedURL.Path, path)
	return
}

// restError returns an error based on status.
func (r *Client) restError(response *http.Response) (err error) {
	status := response.StatusCode
	if status < 400 {
		return
	}
	switch status {
	case http.StatusConflict:
		restError := &Conflict{}
		restError.With(response)
		err = restError
	case http.StatusNotFound:
		restError := &NotFound{}
		restError.With(response)
		err = restError
	default:
		restError := &RestError{}
		restError.With(response)
		err = restError
	}
	return
}

// Field file upload form field.
type Field struct {
	Name     string
	Path     string
	Reader   io.Reader
	Encoding string
}

// Write the field content.
// When Reader is not set, the path is opened and copied.
func (f *Field) Write(writer io.Writer) (err error) {
	if f.Reader == nil {
		file, nErr := os.Open(f.Path)
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		f.Reader = file
		defer func() {
			_ = file.Close()
		}()
	}
	_, err = io.Copy(writer, f.Reader)
	return
}

// encoding returns MIME.
func (f *Field) encoding() (mt string) {
	if f.Encoding != "" {
		mt = f.Encoding
		return
	}
	switch pathlib.Ext(f.Path) {
	case ".json":
		mt = api.MIMEJSON
	case ".yaml":
		mt = api.MIMEYAML
	default:
		mt = "application/octet-stream"
	}
	return
}

// disposition returns content-disposition.
func (f *Field) disposition() (d string) {
	d = fmt.Sprintf(`form-data; name="%s"; filename="%s"`, f.Name, pathlib.Base(f.Path))
	return
}

// refreshToken refreshes the token.
func (r *Client) refreshToken(request *http.Request) (refreshed bool, err error) {
	if r.Login.Token == "" ||
		strings.HasSuffix(request.URL.Path, api.AuthRefreshRoute) {
		return
	}
	login := &api.Login{Refresh: r.Login.Refresh}
	err = r.Post(api.AuthRefreshRoute, login)
	if err == nil {
		r.Login.Token = login.Token
		refreshed = true
		return
	}
	if errors.Is(err, &RestError{}) {
		err = nil
	}
	return
}
