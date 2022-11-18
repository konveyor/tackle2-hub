package addon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/auth"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	pathlib "path"
	"time"
)

const (
	Accept  = "Accept"
	AppJson = "application/json"
)

//
// Param.
type Param struct {
	Key   string
	Value string
}

//
// Client provides a REST client.
type Client struct {
	// baseURL for the nub.
	baseURL string
	// http client.
	http *http.Client
	// addon API token
	token string
	// transport
	transport http.RoundTripper
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
		request.Header.Set(Accept, AppJson)
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
	case http.StatusNotFound:
		err = &NotFound{Path: path}
	default:
		err = errors.New(http.StatusText(status))
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
			Body:   ioutil.NopCloser(reader),
			URL:    r.join(path),
		}
		request.Header.Set(Accept, AppJson)
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
		body, err = ioutil.ReadAll(reply.Body)
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
		err = errors.New(http.StatusText(status))
	}

	return
}

//
// Put a resource.
func (r *Client) Put(path string, object interface{}) (err error) {
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
			Body:   ioutil.NopCloser(reader),
			URL:    r.join(path),
		}
		request.Header.Set(Accept, AppJson)
		return
	}
	reply, err := r.send(request)
	if err != nil {
		return
	}
	status := reply.StatusCode
	switch status {
	case http.StatusNoContent:
	case http.StatusOK:
		var body []byte
		body, err = ioutil.ReadAll(reply.Body)
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
		err = errors.New(http.StatusText(status))
	}

	return
}

//
// Delete a resource.
func (r *Client) Delete(path string) (err error) {
	request := func() (request *http.Request, err error) {
		request = &http.Request{
			Header: http.Header{},
			Method: http.MethodDelete,
			URL:    r.join(path),
		}
		request.Header.Set(Accept, "")
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
		err = errors.New(http.StatusText(status))
	}

	return
}

//
// Send the request.
// Resilient against transient hub availability.
// Retries for 10 minutes.
func (r *Client) send(rb func() (*http.Request, error)) (response *http.Response, err error) {
	var request *http.Request
	err = r.buildTransport()
	if err != nil {
		return
	}
	for i := 0; i < 60; i++ {
		request, err = rb()
		if err != nil {
			return
		}
		request.Header.Set(auth.Header, r.token)
		client := http.Client{Transport: r.transport}
		response, err = client.Do(request)
		if err != nil {
			netErr := &net.OpError{}
			if errors.As(err, &netErr) {
				Log.Info(err.Error())
				time.Sleep(time.Second * 10)
				continue
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
