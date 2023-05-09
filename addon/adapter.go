/*
Tackle hub/addon integration.
*/

package addon

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/settings"
	"golang.org/x/sys/unix"
	"os"
)

var (
	Settings = &settings.Settings
	Log      = logr.WithName("addon")
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
	// TagCategory API.
	TagCategory TagCategory
	// Tag API.
	Tag Tag
	// File API.
	File File
	// RuleBundle API
	RuleBundle RuleBundle
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
			if h.client.Error == nil {
				h.Failed(err.Error())
			}
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
	client := NewClient(Settings.Addon.Hub.URL, Settings.Addon.Hub.Token)
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
		TagCategory: TagCategory{
			client: client,
		},
		Tag: Tag{
			client: client,
		},
		File: File{
			client: client,
		},
		RuleBundle: RuleBundle{
			client: client,
		},
		client: client,
	}

	Log.Info("Addon (adapter) created.")

	return
}
