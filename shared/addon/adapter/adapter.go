/*
Tackle hub/addon integration.
*/

package adapter

import (
	"fmt"
	"os"

	logapi "github.com/go-logr/logr"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"github.com/konveyor/tackle2-hub/shared/task"
	"golang.org/x/sys/unix"
)

var (
	Settings = &settings.Settings.Addon
	Wrap     = liberr.Wrap
	Log      = logr.New("addon", 0)
)

// Addon An addon adapter configured for a task execution.
var Addon *Adapter

func init() {
	unix.Umask(0)
	Addon = New()
}

// Client
type RestClient = binding.RestClient
type RichClient = binding.RichClient
type Params = binding.Params
type Param = binding.Param
type Path = binding.Path

// Errors
type RestError = binding.RestError
type Conflict = binding.Conflict
type NotFound = binding.NotFound

// API namespaces.
type AnalysisProfile = binding.AnalysisProfile
type Application = binding.Application
type Archetype = binding.Archetype
type Bucket = binding.Bucket
type File = binding.File
type Generator = binding.Generator
type Identity = binding.Identity
type Manifest = binding.Manifest
type Platform = binding.Platform
type Proxy = binding.Proxy
type RuleSet = binding.RuleSet
type Schema = binding.Schema
type Setting = binding.Setting
type Tag = binding.Tag
type TagCategory = binding.TagCategory
type Target = binding.Target

// Filter
type Filter = binding.Filter

// The Adapter provides hub/addon integration.
type Adapter struct {
	// Task API.
	Task
	// Log API.
	Log logapi.Logger
	// Wrap error API.
	Wrap func(error, ...any) error
	//
	// AnalysisProfile API.
	AnalysisProfile AnalysisProfile
	// Application API.
	Application Application
	// Archetype
	Archetype Archetype
	// File API.
	File File
	// Generator API.
	Generator Generator
	// Identity API.
	Identity Identity
	// Manifest
	Manifest Manifest
	// Platform
	Platform Platform
	// Proxy API.
	Proxy Proxy
	// RuleSet API
	RuleSet RuleSet
	// Schema API
	Schema Schema
	// Settings API.
	Setting Setting
	// TagCategory API.
	TagCategory TagCategory
	// Tag API.
	Tag Tag
	// Target API
	Target Target
}

// Run addon.
// Reports:
//   - Started
//   - Succeeded
//   - Failed (when addon returns error).
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
				err = fmt.Errorf("%#v", r)
			}
		}
		if err != nil {
			h.Log.Error(err, "Addon failed.")
			h.Failed(err.Error())
			os.Exit(1)
		}
	}()
	//
	// Report addon started.
	h.Load()
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

// Use sets the richClient.
func (h *Adapter) Use(richClient *RichClient) {
	h.Log = Log
	h.Wrap = Wrap
	h.richClient = richClient
	h.AnalysisProfile = richClient.AnalysisProfile
	h.Application = richClient.Application
	h.Archetype = richClient.Archetype
	h.File = richClient.File
	h.Generator = richClient.Generator
	h.Identity = richClient.Identity
	h.Manifest = richClient.Manifest
	h.Platform = richClient.Platform
	h.Proxy = richClient.Proxy
	h.RuleSet = richClient.RuleSet
	h.Schema = richClient.Schema
	h.Setting = richClient.Setting
	h.Tag = richClient.Tag
	h.TagCategory = richClient.TagCategory
	h.Target = richClient.Target
}

// Client returns the rich-client.
func (h *Adapter) Client() (richClient *RichClient) {
	richClient = h.richClient
	return
}

// New builds a new Addon Adapter object.
func New() (adapter *Adapter) {
	richClient := binding.New(Settings.Hub.URL)
	richClient.Client.Use(api.Login{Token: Settings.Hub.Token})
	adapter = &Adapter{}
	adapter.Use(richClient)
	Log.Info("Addon (adapter) created.")
	return
}
