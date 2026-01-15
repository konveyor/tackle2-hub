package model

import (
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	"github.com/konveyor/tackle2-hub/internal/migration/v21/model"
)

// Field (data) types.
type JSON = model.JSON

var ALL = model.All()

// Models
type Model = model.Model
type Application = model.Application
type Archetype = model.Archetype
type Assessment = model.Assessment
type TechDependency = model.TechDependency
type Incident = model.Incident
type Analysis = model.Analysis
type AnalysisProfile = model.AnalysisProfile
type Insight = model.Insight
type Bucket = model.Bucket
type BucketOwner = model.BucketOwner
type BusinessService = model.BusinessService
type Dependency = model.Dependency
type File = model.File
type Fact = model.Fact
type Generator = model.Generator
type Identity = model.Identity
type Import = model.Import
type ImportSummary = model.ImportSummary
type ImportTag = model.ImportTag
type JobFunction = model.JobFunction
type Manifest = model.Manifest
type MigrationWave = model.MigrationWave
type PK = model.PK
type Platform = model.Platform
type ProfileGenerator = model.ProfileGenerator
type Proxy = model.Proxy
type Questionnaire = model.Questionnaire
type Review = model.Review
type Setting = model.Setting
type RuleSet = model.RuleSet
type Rule = model.Rule
type Stakeholder = model.Stakeholder
type StakeholderGroup = model.StakeholderGroup
type Tag = model.Tag
type TagCategory = model.TagCategory
type Target = model.Target
type TargetProfile = model.TargetProfile
type Task = model.Task
type TaskGroup = model.TaskGroup
type TaskReport = model.TaskReport
type Ticket = model.Ticket
type Tracker = model.Tracker

// JSON fields
type Ref = json.Ref
type Map = json.Map
type Data = json.Data
type Document = json.Document
type ArchivedInsight = model.ArchivedInsight
type Attachment = model.Attachment
type Link = model.Link
type Repository = model.Repository
type TargetLabel = model.TargetLabel
type TaskError = model.TaskError
type TaskEvent = model.TaskEvent
type TaskPolicy = model.TaskPolicy
type TTL = model.TTL
type InExList = model.InExList
type TargetSelection = model.TargetSelection

// Assessment JSON fields
type Section = model.Section
type Question = model.Question
type Answer = model.Answer
type Thresholds = model.Thresholds
type RiskMessages = model.RiskMessages
type CategorizedTag = model.CategorizedTag

// Join tables
type ApplicationTag = model.ApplicationTag
type ApplicationIdentity = model.ApplicationIdentity

// Errors
type DependencyCyclicError = model.DependencyCyclicError
