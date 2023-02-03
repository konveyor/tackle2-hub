package model

import (
	"github.com/konveyor/tackle2-hub/migration/v3/model"
	"gorm.io/datatypes"
)

//
// Field (data) types.
type JSON = datatypes.JSON

//
// Models
type Model = model.Model
type Application = model.Application
type BucketOwner = model.BucketOwner
type BusinessService = model.BusinessService
type Dependency = model.Dependency
type File = model.File
type Identity = model.Identity
type Import = model.Import
type ImportSummary = model.ImportSummary
type ImportTag = model.ImportTag
type JobFunction = model.JobFunction
type Proxy = model.Proxy
type Review = model.Review
type Setting = model.Setting
type Stakeholder = model.Stakeholder
type StakeholderGroup = model.StakeholderGroup
type Tag = model.Tag
type TagType = model.TagType
type Task = model.Task
type TaskGroup = model.TaskGroup
type TaskReport = model.TaskReport
type Ticket = model.Ticket
type Tracker = model.Tracker
type TTL = model.TTL
type Metadata = model.Metadata
type Project = model.Project
type IssueType = model.IssueType

//
// Errors
type DependencyCyclicError = model.DependencyCyclicError
