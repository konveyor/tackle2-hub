package binding

import (
	"github.com/konveyor/tackle2-hub/shared/binding/analysis"
	"github.com/konveyor/tackle2-hub/shared/binding/application"
	"github.com/konveyor/tackle2-hub/shared/binding/archetype"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	_import "github.com/konveyor/tackle2-hub/shared/binding/import"
	"github.com/konveyor/tackle2-hub/shared/binding/report"
	"github.com/konveyor/tackle2-hub/shared/binding/task"
	"github.com/konveyor/tackle2-hub/shared/binding/taskgroup"
)

// Type aliases for subpackage types.
type (
	Analysis    = analysis.Analysis
	Application = application.Application
	Archetype   = archetype.Archetype
	Bucket      = bucket.Bucket
	Import      = _import.Import
	Report      = report.Report
	Task        = task.Task
	TaskGroup   = taskgroup.TaskGroup
)
