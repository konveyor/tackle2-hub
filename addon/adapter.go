/*
Tackle hub/addon integration.
*/

package addon

import (
	"fmt"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/settings"
	"golang.org/x/sys/unix"
	"net/http"
	"os"
	"strings"
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("addon")
)

//
// Addon An addon adapter configured for a task execution.
var Addon *Adapter

func init() {
	unix.Umask(0)
	err := Settings.Load()
	if err != nil {
		panic(err)
	}

	Addon = newAdapter()
}

//
// The Adapter provides hub/addon integration.
type Adapter struct {
	// Task API.
	Task
	// Settings API
	Setting Setting
	// Application API.
	Application Application
	// Identity API.
	Identity Identity
	// Proxy API.
	Proxy Proxy
	// TagType API.
	TagType TagType
	// Tag API.
	Tag Tag
	// File API.
	File File
	// client A REST client.
	client *Client
}

//
// Run addon.
// Reports:
//  - Started
//  - Succeeded
//  - Failed (when addon returns error).
func (h *Adapter) Run(addon func() error) {
	var err error
	//
	// Error handling.
	defer func() {
		r := recover()
		if r != nil {
			if pErr, cast := r.(error); cast {
				err = pErr
			} else {
				panic(r)
			}
		}
		if err != nil {
			if _, soft := err.(interface{ Soft() *SoftError }); !soft {
				Log.Error(err, "Addon failed.")
			}
			h.Failed(err.Error())
			os.Exit(1)
		}
	}()
	//
	// Report addon started.
	h.Started()
	//
	// Run addon.
	err = addon()
	if err != nil {
		return
	}
	//
	// Report addon succeeded.
	h.Succeeded()
}

//
// Client provides the REST client.
func (h *Adapter) Client() *Client {
	return h.client
}

//
// newAdapter builds a new Addon Adapter object.
func newAdapter() (adapter *Adapter) {
	//
	// Build REST client.
	client := &Client{
		baseURL: Settings.Addon.Hub.URL,
		http:    &http.Client{},
		token:   Settings.Addon.Hub.Token,
	}
	//
	// Build Adapter.
	adapter = &Adapter{
		Task: Task{
			client: client,
		},
		Setting: Setting{
			client: client,
		},
		Application: Application{
			client: client,
		},
		Identity: Identity{
			client: client,
		},
		Proxy: Proxy{
			client: client,
		},
		TagType: TagType{
			client: client,
		},
		Tag: Tag{
			client: client,
		},
		File: File{
			client: client,
		},
		client: client,
	}

	Log.Info("Addon (adapter) created.")

	return
}

//
// Params mapping.
type Params map[string]interface{}

//
// inject values into path.
func (p Params) inject(path string) (s string) {
	in := strings.Split(path, "/")
	for i := range in {
		if len(in[i]) < 1 {
			continue
		}
		key := in[i][1:]
		if v, found := p[key]; found {
			in[i] = fmt.Sprintf("%v", v)
		}
	}
	s = strings.Join(in, "/")
	return
}
