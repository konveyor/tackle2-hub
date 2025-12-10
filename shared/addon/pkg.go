package addon

import (
	"github.com/konveyor/tackle2-hub/shared/addon/adapter"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Addon = adapter.Addon
	Log   = adapter.Log
)

// Environment.
const (
	EnvSharedDir = settings.EnvSharedPath
	EnvCacheDir  = settings.EnvCachePath
	EnvToken     = settings.EnvHubToken
	EnvTask      = settings.EnvTask
)

// Client
type Client = adapter.Client
type Params = adapter.Params
type Param = adapter.Param
type Path = adapter.Path

// Errors
type ResetError = adapter.RestError
type Conflict = adapter.Conflict
type NotFound = adapter.NotFound

// Handlers
type Application = adapter.Application
type Bucket = adapter.Bucket
type BucketContent = adapter.BucketContent
type File = adapter.File
type Identity = adapter.Identity
type Manifest = adapter.Manifest
type Platform = adapter.Platform
type Proxy = adapter.Proxy
type RuleSet = adapter.RuleSet
type Schema = adapter.Schema
type Setting = adapter.Setting
type Tag = adapter.Tag
type TagCategory = adapter.TagCategory
type Archetype = adapter.Archetype
type Generator = adapter.Generator
