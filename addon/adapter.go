/*
Tackle hub/addon integration.
*/

package addon

import (
	"fmt"
	"os"

	logapi "github.com/go-logr/logr"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"golang.org/x/sys/unix"
)

var (
	Settings = &settings.Settings
	Wrap     = liberr.Wrap
	Log      = logr.WithName("addon")
)

// Environment.
const (
	EnvSharedDir = settings.EnvSharedPath
	EnvCacheDir  = settings.EnvCachePath
	EnvToken     = settings.EnvHubToken
	EnvTask      = settings.EnvTask
)

// Addon An addon adapter configured for a task execution.
var Addon *Adapter

func init() {
	unix.Umask(0)
	err := Settings.Addon.Load()
	if err != nil {
		panic(err)
	}

	Addon = newAdapter()
}

// Client
type Client = binding.Client
type Params = binding.Params
type Param = binding.Param
type Path = binding.Path

// Error
type ResetError = binding.RestError
type Conflict = binding.Conflict
type NotFound = binding.NotFound

// Handler
type Application = binding.Application
type Bucket = binding.Bucket
type BucketContent = binding.BucketContent
type File = binding.File
type Identity = binding.Identity
type Manifest = binding.Manifest
type Platform = binding.Platform
type Proxy = binding.Proxy
type RuleSet = binding.RuleSet
type Schema = binding.Schema
type Setting = binding.Setting
type Tag = binding.Tag
type TagCategory = binding.TagCategory
type Archetype = binding.Archetype
type Generator = binding.Generator

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
	// Settings API.
	Setting Setting
	// Schema API
	Schema Schema
	// Application API.
	Application Application
	// Identity API.
	Identity Identity
	// Manifest
	Manifest Manifest
	// Platform
	Platform Platform
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
	// Generator API.
	Generator Generator
	// Archetype
	Archetype Archetype
	// client A REST client.
	client *Client
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

// newAdapter builds a new Addon Adapter object.
func newAdapter() (adapter *Adapter) {
	richClient := binding.New(Settings.Addon.Hub.URL)
	richClient.Client.Login.Token = Settings.Addon.Hub.Token
	adapter = &Adapter{
		client: richClient.Client,
		Task: Task{
			richClient: richClient,
		},
		Log:         Log,
		Wrap:        Wrap,
		Setting:     richClient.Setting,
		Schema:      richClient.Schema,
		Application: richClient.Application,
		Identity:    richClient.Identity,
		Manifest:    richClient.Manifest,
		Platform:    richClient.Platform,
		Proxy:       richClient.Proxy,
		TagCategory: richClient.TagCategory,
		Tag:         richClient.Tag,
		File:        richClient.File,
		RuleSet:     richClient.RuleSet,
		Generator:   richClient.Generator,
		Archetype:   richClient.Archetype,
	}

	Log.Info("Addon (adapter) created.")

	return
}
