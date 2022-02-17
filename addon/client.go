package addon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
}

//
// Get a resource.
func (r *Client) Get(path string, object interface{}) (err error) {
	request := &http.Request{
		Method: http.MethodGet,
		URL:    r.join(path),
	}
	reply, err := r.http.Do(request)
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
			return
		}
		err = json.Unmarshal(body, object)
	case http.StatusNotFound:
		err = &NotFound{path}
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
		return
	}
	reader := bytes.NewReader(bfr)
	request := &http.Request{
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(reader),
		URL:    r.join(path),
	}
	reply, err := r.http.Do(request)
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
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			return
		}
	case http.StatusConflict:
		err = &Conflict{path}
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
		return
	}
	reader := bytes.NewReader(bfr)
	request := &http.Request{
		Method: http.MethodPut,
		Body:   ioutil.NopCloser(reader),
		URL:    r.join(path),
	}
	reply, err := r.http.Do(request)
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
			return
		}
		err = json.Unmarshal(body, object)
		if err != nil {
			return
		}
	case http.StatusNotFound:
		err = &NotFound{path}
	default:
		err = errors.New(http.StatusText(status))
	}

	return
}

//
// Delete a resource.
func (r *Client) Delete(path string) (err error) {
	request := &http.Request{
		Method: http.MethodDelete,
		URL:    r.join(path),
	}
	reply, err := r.http.Do(request)
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
		err = &NotFound{path}
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

//
// Conflict reports 409 error.
type Conflict struct {
	Path string
}

func (e Conflict) Error() string {
	return fmt.Sprintf("POST: path:%s [conflict]", e.Path)
}

func (e *Conflict) Is(err error) (matched bool) {
	_, matched = err.(*Conflict)
	return
}

//
// NotFound reports 404 error.
type NotFound struct {
	Path string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("GET: path:%s [not-found]", e.Path)
}

func (e *NotFound) Is(err error) (matched bool) {
	_, matched = err.(*NotFound)
	return
}
