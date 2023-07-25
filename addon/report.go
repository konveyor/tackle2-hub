package addon

import (
	"fmt"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/task"
)

//
// TaskReport API.
type TaskReport struct {
	// Task
	task int
	// Task report.
	report api.TaskReport
	// Client
	client *binding.Client
}

//
// Started report addon started.
func (h *TaskReport) Started() {
	h.delete()
	h.report.Status = task.Running
	h.Push()
	Log.Info("Addon reported started.")
	return
}

//
// Succeeded report addon succeeded.
func (h *TaskReport) Succeeded() {
	h.report.Status = task.Succeeded
	h.report.Completed = h.report.Total
	h.Push()
	Log.Info("Addon reported: succeeded.")
	return
}

//
// Failed report addon failed.
// The reason can be a printf style format.
func (h *TaskReport) Failed(reason string, x ...interface{}) {
	reason = fmt.Sprintf(reason, x...)
	h.report.Status = task.Failed
	h.report.Errors = append(
		h.report.Errors,
		api.TaskError{
			Severity:    "Error",
			Description: reason,
		})
	h.Push()
	Log.Info(
		"Addon reported: failed.",
		"reason",
		reason)
	return
}

//
// Error report addon error.
// The description can be a printf style format.
func (h *TaskReport) Error(severity, description string, x ...interface{}) {
	h.report.Status = task.Failed
	description = fmt.Sprintf(description, x...)
	h.report.Errors = append(
		h.report.Errors,
		api.TaskError{
			Severity:    severity,
			Description: description,
		})
	h.Push()
	return
}

//
// Activity report addon activity.
// The description can be a printf style format.
func (h *TaskReport) Activity(entry string, x ...interface{}) {
	entry = fmt.Sprintf(entry, x...)
	h.report.Activity = append(
		h.report.Activity,
		entry)
	h.Push()
	Log.Info(
		"Addon reported: activity.",
		"activity",
		h.report.Activity)
	return
}

//
// Total report addon total items.
func (h *TaskReport) Total(n int) {
	h.report.Total = n
	h.Push()
	Log.Info(
		"Addon updated: total.",
		"total",
		h.report.Total)
	return
}

//
// Increment report addon completed (+1) items.
func (h *TaskReport) Increment() {
	h.report.Completed++
	h.Push()
	Log.Info(
		"Addon updated: total.",
		"total",
		h.report.Total)
	return
}

//
// Completed report addon completed (N) items.
func (h *TaskReport) Completed(n int) {
	h.report.Completed = n
	h.Push()
	Log.Info("Addon reported: completed.")
	return
}

//
// Result report addon result.
func (h *TaskReport) Result(object interface{}) {
	h.report.Result = object
	h.Push()
	Log.Info("Addon reported: result.")
	return
}

//
// Push create/update the task report.
func (h *TaskReport) Push() {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	params := Params{
		api.ID: h.task,
	}
	path := Path(api.TaskReportRoot).Inject(params)
	if h.report.ID == 0 {
		err = h.client.Post(path, &h.report)
	} else {
		err = h.client.Put(path, &h.report)
	}
	return
}

//
// delete deletes the task report.
func (h *TaskReport) delete() {
	params := Params{
		api.ID: h.task,
	}
	path := Path(api.TaskReportRoot).Inject(params)
	err := h.client.Delete(path)
	if err != nil {
		panic(err)
	}
}
