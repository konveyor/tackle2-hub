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
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"golang.org/x/sys/unix"
)

var (
	Settings = &settings.Settings
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
type Client = binding2.Client
type Params = binding2.Params
type Param = binding2.Param
type Path = binding2.Path

// Errors
type RestError = binding2.RestError
type Conflict = binding2.Conflict
type NotFound = binding2.NotFound

// Handlers
type Application = binding2.Application
type Bucket = binding2.Bucket
type BucketContent = binding2.BucketContent
type File = binding2.File
type Identity = binding2.Identity
type Manifest = binding2.Manifest
type Platform = binding2.Platform
type Proxy = binding2.Proxy
type RuleSet = binding2.RuleSet
type Schema = binding2.Schema
type Setting = binding2.Setting
type Tag = binding2.Tag
type TagCategory = binding2.TagCategory
type Archetype = binding2.Archetype
type Generator = binding2.Generator

// Filter
type Filter = binding2.Filter

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
	// SCM
	SCM SCM
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

// New builds a new Addon Adapter object.
func New() (adapter *Adapter) {
	richClient := binding2.New(Settings.Hub.URL)
	richClient.Client.Login.Token = Settings.Hub.Token
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
		SCM:         SCM{},
	}

	Log.Info("Addon (adapter) created.")

	return
}
