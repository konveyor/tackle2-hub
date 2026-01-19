package client

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// RestError reports REST errors.
type RestError struct {
	Reason string
	Method string
	Path   string
	Status int
	Body   string
}

func (e *RestError) Is(err error) (matched bool) {
	var inst *RestError
	matched = errors.As(err, &inst)
	return
}

func (e *RestError) Error() (s string) {
	s = e.Reason
	return
}

func (e *RestError) With(r *http.Response) {
	e.Method = r.Request.Method
	e.Path = r.Request.URL.Path
	e.Status = r.StatusCode
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			e.Body = string(body)
		}
	}
	s := strings.ToUpper(e.Method)
	s += " "
	s += e.Path
	s += " failed: "
	s += strconv.Itoa(e.Status)
	s += "("
	s += http.StatusText(e.Status)
	s += ")"
	if e.Body != "" {
		s += " body: "
		s += e.Body
	}
	e.Reason = s
}

// Conflict reports 409 error.
type Conflict struct {
	RestError
}

func (e *Conflict) Is(err error) (matched bool) {
	var inst *Conflict
	matched = errors.As(err, &inst)
	return
}

// NotFound reports 404 error.
type NotFound struct {
	RestError
}

func (e *NotFound) Is(err error) (matched bool) {
	var inst *NotFound
	matched = errors.As(err, &inst)
	return
}

// EmptyBody reports an empty body.
type EmptyBody struct {
	RestError
}

func (e *EmptyBody) Is(err error) (matched bool) {
	var inst *EmptyBody
	matched = errors.As(err, &inst)
	return
}
