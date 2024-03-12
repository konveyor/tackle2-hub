package addon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/task"
)

// Task API.
type Task struct {
	richClient *binding.RichClient
	// Task
	task *api.Task
	// Task report.
	report api.TaskReport
}

// Load a task by ID.
func (h *Task) Load() {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	h.task, err = h.richClient.Task.Get(uint(Settings.Addon.Task))
	return
}

// Application returns the application associated with the task.
func (h *Task) Application() (r *api.Application, err error) {
	appRef := h.task.Application
	if appRef == nil {
		err = &NotFound{}
		return
	}
	r, err = h.richClient.Application.Get(appRef.ID)
	return
}

// Data returns the addon data.
func (h *Task) Data() (d map[string]interface{}) {
	d = h.task.Data.(map[string]interface{})
	return
}

// DataWith populates the addon data object.
func (h *Task) DataWith(object interface{}) (err error) {
	b, _ := json.Marshal(h.task.Data)
	err = json.Unmarshal(b, object)
	return
}

// Started report addon started.
func (h *Task) Started() {
	h.deleteReport()
	h.report.Status = task.Running
	h.pushReport()
	Log.Info("Addon reported started.")
	return
}

// Succeeded report addon succeeded.
func (h *Task) Succeeded() {
	h.report.Status = task.Succeeded
	h.report.Completed = h.report.Total
	h.pushReport()
	Log.Info("Addon reported: succeeded.")
	return
}

// Failed report addon failed.
// The reason can be a printf style format.
func (h *Task) Failed(reason string, v ...interface{}) {
	reason = fmt.Sprintf(reason, v...)
	h.Error(api.TaskError{
		Severity:    "Error",
		Description: reason,
	})
	Log.Info(
		"Addon reported: failed.",
		"reason",
		reason)
	return
}

// Errorf report addon error.
func (h *Task) Errorf(severity, description string, v ...interface{}) {
	h.Error(api.TaskError{
		Severity:    severity,
		Description: fmt.Sprintf(description, v...),
	})
}

// Error report addon error.
func (h *Task) Error(error ...api.TaskError) {
	h.report.Status = task.Failed
	for i := range error {
		h.report.Errors = append(
			h.report.Errors,
			error[i])
		Log.Info(
			"Addon reported: error.",
			"error",
			error[i])
	}
	h.pushReport()
	return
}

// Activity report addon activity.
// The description can be a printf style format.
func (h *Task) Activity(entry string, v ...interface{}) {
	entry = fmt.Sprintf(entry, v...)
	lines := strings.Split(entry, "\n")
	for i := range lines {
		if i > 0 {
			entry = "> " + lines[i]
		} else {
			entry = lines[i]
		}
		h.report.Activity = append(
			h.report.Activity,
			entry)
		Log.Info(
			"Addon reported: activity.",
			"entry",
			entry)
	}
	h.pushReport()
	return
}

// Attach ensures the file is attached to the report
// associated with the last entry in the activity.
func (h *Task) Attach(f *api.File) {
	index := len(h.report.Activity)
	h.AttachAt(f, index)
	return
}

// AttachAt ensures the file is attached to
// the report indexed to the activity.
// The activity is a 1-based index. Zero(0) means NOT associated.
func (h *Task) AttachAt(f *api.File, activity int) {
	for _, ref := range h.report.Attached {
		if ref.ID == f.ID {
			return
		}
	}
	Log.Info(
		"Addon attached.",
		"ID",
		f.ID,
		"name",
		f.Name,
		"activity",
		activity)
	h.report.Attached = append(
		h.report.Attached,
		api.Attachment{
			Activity: activity,
			Ref: api.Ref{
				ID:   f.ID,
				Name: f.Name,
			},
		})
	h.pushReport()
	return
}

// Total report addon total items.
func (h *Task) Total(n int) {
	h.report.Total = n
	h.pushReport()
	Log.Info(
		"Addon updated: total.",
		"total",
		h.report.Total)
	return
}

// Increment report addon completed (+1) items.
func (h *Task) Increment() {
	h.report.Completed++
	h.pushReport()
	Log.Info(
		"Addon updated: total.",
		"total",
		h.report.Total)
	return
}

// Completed report addon completed (N) items.
func (h *Task) Completed(n int) {
	h.report.Completed = n
	h.pushReport()
	Log.Info("Addon reported: completed.")
	return
}

// Bucket returns the bucket API.
func (h *Task) Bucket() (b *binding.BucketContent) {
	b = h.richClient.Task.Bucket(h.task.ID)
	return
}

// Result report addon result.
func (h *Task) Result(object interface{}) {
	h.report.Result = object
	h.pushReport()
	Log.Info("Addon reported: result.")
	return
}

// deleteReport deletes the task report.
func (h *Task) deleteReport() {
	params := Params{
		api.ID: h.task.ID,
	}
	path := Path(api.TaskReportRoot).Inject(params)
	err := h.richClient.Client.Delete(path)
	if err != nil {
		panic(err)
	}
}

// pushReport create/update the task report.
func (h *Task) pushReport() {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	params := Params{
		api.ID: h.task.ID,
	}
	client := h.richClient.Client
	path := Path(api.TaskReportRoot).Inject(params)
	if h.report.ID == 0 {
		err = client.Post(path, &h.report)
	} else {
		err = client.Put(path, &h.report)
	}
	return
}
