package model

import (
	"github.com/konveyor/tackle2-hub/migration/v5/model"
	"gorm.io/datatypes"
)

//
// Field (data) types.
type JSON = datatypes.JSON

//
// Models
type Model = model.Model
type Application = model.Application
type TechDependency = model.TechDependency
type Incident = model.Incident
type Analysis = model.Analysis
type Issue = model.Issue
type Bucket = model.Bucket
type BucketOwner = model.BucketOwner
type BusinessService = model.BusinessService
type Dependency = model.Dependency
type File = model.File
type Fact = model.Fact
type Identity = model.Identity
type Import = model.Import
type ImportSummary = model.ImportSummary
type ImportTag = model.ImportTag
type JobFunction = model.JobFunction
type MigrationWave = model.MigrationWave
type Proxy = model.Proxy
type Review = model.Review
type Setting = model.Setting
type RuleSet = model.RuleSet
type Rule = model.Rule
type Stakeholder = model.Stakeholder
type StakeholderGroup = model.StakeholderGroup
type Tag = model.Tag
type TagCategory = model.TagCategory
type Task = model.Task
type TaskGroup = model.TaskGroup
type TaskReport = model.TaskReport
type Ticket = model.Ticket
type Tracker = model.Tracker

//
type TTL = model.TTL

//
// Join tables
type ApplicationTag = model.ApplicationTag

//
// Errors
type DependencyCyclicError = model.DependencyCyclicError
