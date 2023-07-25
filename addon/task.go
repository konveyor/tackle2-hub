package addon

import (
	"encoding/json"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
)

//
// Task API.
type Task struct {
	// Client
	richClient *binding.RichClient
	// Task
	task *api.Task
	// Task report
	Report TaskReport
}

//
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

//
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

//
// Data returns the addon data.
func (h *Task) Data() (d map[string]interface{}) {
	d = h.task.Data.(map[string]interface{})
	return
}

//
// DataWith populates the addon data object.
func (h *Task) DataWith(object interface{}) (err error) {
	b, _ := json.Marshal(h.task.Data)
	err = json.Unmarshal(b, object)
	return
}

//
// Variant returns the task variant.
func (h *Task) Variant() string {
	return h.task.Variant
}

//
// Bucket returns the bucket API.
func (h *Task) Bucket() (b *binding.BucketContent) {
	b = h.richClient.Task.Bucket(h.task.ID)
	return
}

//
// Started report addon started.
func (h *Task) Started() {
	h.Report.Started()
	return
}

//
// Succeeded report addon succeeded.
func (h *Task) Succeeded() {
	h.Report.Succeeded()
	h.Report.Push()
	return
}

//
// Failed report addon failed.
// The reason can be a printf style format.
func (h *Task) Failed(reason string, x ...interface{}) {
	h.Report.Failed(reason, x...)
	h.Report.Push()
}

//
// Error report addon error.
// The description can be a printf style format.
func (h *Task) Error(severity, description string, x ...interface{}) {
	h.Report.Error(severity, description, x...)
	h.Report.Push()
	return
}

//
// Activity report addon activity.
// The description can be a printf style format.
func (h *Task) Activity(entry string, x ...interface{}) {
	h.Report.Activity(entry, x...)
	h.Report.Push()
}

//
// Total report addon total items.
func (h *Task) Total(n int) {
	h.Report.Total(n)
	h.Report.Push()
}

//
// Increment report addon completed (+1) items.
func (h *Task) Increment() {
	h.Report.Increment()
	h.Report.Push()
}

//
// Completed report addon completed (N) items.
func (h *Task) Completed(n int) {
	h.Report.Completed(n)
	h.Report.Push()
}

//
// Result report addon result.
func (h *Task) Result(object interface{}) {
	h.Report.Result(object)
	h.Report.Push()
}
