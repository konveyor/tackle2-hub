package addon

import (
	"bytes"
	"encoding/json"
	"errors"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/auth"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

//
// Client provides a REST client.
type Client struct {
	// baseURL for the nub.
	baseURL string
	// http client.
	http *http.Client
	// addon API token
	token string
}

//
// Get a resource.
func (r *Client) Get(path string, object interface{}) (err error) {
	request := &http.Request{
		Header: http.Header{},
		Method: http.MethodGet,
		URL:    r.join(path),
	}
	request.Header.Set(auth.Header, r.token)
	reply, err := r.http.Do(request)
	if err != nil {
		err = liberr.Wrap(err)
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
	bfr, err := json.Marshal(object)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	reader := bytes.NewReader(bfr)
	request := &http.Request{
		Header: http.Header{},
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(reader),
		URL:    r.join(path),
	}
	request.Header.Set(auth.Header, r.token)
	reply, err := r.http.Do(request)
	if err != nil {
		err = liberr.Wrap(err)
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
	bfr, err := json.Marshal(object)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	reader := bytes.NewReader(bfr)
	request := &http.Request{
		Header: http.Header{},
		Method: http.MethodPut,
		Body:   ioutil.NopCloser(reader),
		URL:    r.join(path),
	}
	request.Header.Set(auth.Header, r.token)
	reply, err := r.http.Do(request)
	if err != nil {
		err = liberr.Wrap(err)
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
	request := &http.Request{
		Header: http.Header{},
		Method: http.MethodDelete,
		URL:    r.join(path),
	}
	request.Header.Set(auth.Header, r.token)
	reply, err := r.http.Do(request)
	if err != nil {
		err = liberr.Wrap(err)
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

func (r *Client) join(path string) (parsedURL *url.URL) {
	parsedURL, _ = url.Parse(r.baseURL)
	parsedURL.Path = path
	return
}
