/*
Tackle hub/addon integration.
*/

package addon

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
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
// Client
type Client = binding.Client
type Params = binding.Params
type Param = binding.Param
type Path = binding.Path
type Field = binding.Field

//
// Error
type SoftError = binding.SoftError
type ResetError = binding.RestError
type Conflict = binding.Conflict
type NotFound = binding.NotFound

//
// Handler
type Application = binding.Application
type Bucket = binding.Bucket
type BucketContent = binding.BucketContent
type File = binding.File
type Identity = binding.Identity
type Proxy = binding.Proxy
type RuleSet = binding.RuleSet
type Setting = binding.Setting
type Tag = binding.Tag
type TagCategory = binding.TagCategory

//
// The Adapter provides hub/addon integration.
type Adapter struct {
	// Task API.
	Task
	// Settings API.
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
	// RuleSet API
	RuleSet RuleSet
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
				os.Exit(1)
			}
			h.Failed(err.Error())
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
	switch h.report.Status {
	case task.Failed,
		task.Succeeded:
	default:
		h.Succeeded()
	}
}

//
// newAdapter builds a new Addon Adapter object.
func newAdapter() (adapter *Adapter) {
	richClient := binding.New(Settings.Addon.Hub.URL)
	richClient.Client.SetToken(Settings.Addon.Hub.Token)
	adapter = &Adapter{
		client: richClient.Client,
		Task: Task{
			richClient: richClient,
		},
		Setting:     richClient.Setting,
		Application: richClient.Application,
		Identity:    richClient.Identity,
		Proxy:       richClient.Proxy,
		TagCategory: richClient.TagCategory,
		Tag:         richClient.Tag,
		File:        richClient.File,
		RuleSet:     richClient.RuleSet,
	}

	Log.Info("Addon (adapter) created.")

	return
}
