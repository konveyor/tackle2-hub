package model

import (
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	model2 "github.com/konveyor/tackle2-hub/internal/migration/v21/model"
)

// Field (data) types.
type JSON = model2.JSON

var ALL = model2.All()

// Models
type Model = model2.Model
type Application = model2.Application
type Archetype = model2.Archetype
type Assessment = model2.Assessment
type TechDependency = model2.TechDependency
type Incident = model2.Incident
type Analysis = model2.Analysis
type AnalysisProfile = model2.AnalysisProfile
type Insight = model2.Insight
type Bucket = model2.Bucket
type BucketOwner = model2.BucketOwner
type BusinessService = model2.BusinessService
type Dependency = model2.Dependency
type File = model2.File
type Fact = model2.Fact
type Generator = model2.Generator
type Identity = model2.Identity
type Import = model2.Import
type ImportSummary = model2.ImportSummary
type ImportTag = model2.ImportTag
type JobFunction = model2.JobFunction
type Manifest = model2.Manifest
type MigrationWave = model2.MigrationWave
type PK = model2.PK
type Platform = model2.Platform
type ProfileGenerator = model2.ProfileGenerator
type Proxy = model2.Proxy
type Questionnaire = model2.Questionnaire
type Review = model2.Review
type Setting = model2.Setting
type RuleSet = model2.RuleSet
type Rule = model2.Rule
type Stakeholder = model2.Stakeholder
type StakeholderGroup = model2.StakeholderGroup
type Tag = model2.Tag
type TagCategory = model2.TagCategory
type Target = model2.Target
type TargetProfile = model2.TargetProfile
type Task = model2.Task
type TaskGroup = model2.TaskGroup
type TaskReport = model2.TaskReport
type Ticket = model2.Ticket
type Tracker = model2.Tracker

// JSON fields
type Ref = json.Ref
type Map = json.Map
type Data = json.Data
type Document = json.Document
type ArchivedInsight = model2.ArchivedInsight
type Attachment = model2.Attachment
type Link = model2.Link
type Repository = model2.Repository
type TargetLabel = model2.TargetLabel
type TaskError = model2.TaskError
type TaskEvent = model2.TaskEvent
type TaskPolicy = model2.TaskPolicy
type TTL = model2.TTL
type InExList = model2.InExList

// Assessment JSON fields
type Section = model2.Section
type Question = model2.Question
type Answer = model2.Answer
type Thresholds = model2.Thresholds
type RiskMessages = model2.RiskMessages
type CategorizedTag = model2.CategorizedTag

// Join tables
type ApplicationTag = model2.ApplicationTag
type ApplicationIdentity = model2.ApplicationIdentity

// Errors
type DependencyCyclicError = model2.DependencyCyclicError
