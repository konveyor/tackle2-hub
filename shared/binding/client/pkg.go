package client

import (
	"fmt"
	"strings"
	"time"

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
