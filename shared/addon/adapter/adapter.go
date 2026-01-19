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
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/application"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
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
type Client = binding.Client
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
type Application = application.Application
type Archetype = binding.Archetype
type Bucket = bucket.Bucket
type BucketContent = bucket.Content
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

// Client returns a configured rich-client.
func (h *Adapter) Client() (client *RichClient) {
	client = &RichClient{}
	client.Client.Login = h.Task.richClient.Client.Login
	client.Client.Retry = h.Task.richClient.Client.Retry
	return
}

// New builds a new Addon Adapter object.
func New() (adapter *Adapter) {
	richClient := binding.New(Settings.Hub.URL)
	richClient.Client.Login.Token = Settings.Hub.Token
	adapter = &Adapter{
		Task: Task{
			richClient: richClient,
		},
		Log:  Log,
		Wrap: Wrap,
		//
		AnalysisProfile: richClient.AnalysisProfile,
		Application:     richClient.Application,
		Archetype:       richClient.Archetype,
		File:            richClient.File,
		Generator:       richClient.Generator,
		Identity:        richClient.Identity,
		Manifest:        richClient.Manifest,
		Platform:        richClient.Platform,
		Proxy:           richClient.Proxy,
		RuleSet:         richClient.RuleSet,
		Schema:          richClient.Schema,
		Setting:         richClient.Setting,
		Tag:             richClient.Tag,
		TagCategory:     richClient.TagCategory,
		Target:          richClient.Target,
	}

	Log.Info("Addon (adapter) created.")

	return
}
