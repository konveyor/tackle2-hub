package report

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) (h Report) {
	h.Task = Task{client: client}
	return
}

// Report API.
type Report struct {
	Task Task
}

// Task report API.
type Task struct {
	client client.RestClient
}

// Queued returns queued task report.
func (h Task) Queued() (r *api.TaskQueue, err error) {
	r = &api.TaskQueue{}
	err = h.client.Get(api.TasksReportQueueRoute, r)
	return
}

// Dashboard returns task dashboard report.
func (h Task) Dashboard() (list []api.TaskDashboard, err error) {
	list = []api.TaskDashboard{}
	err = h.client.Get(api.TasksReportDashboardRoute, &list)
	return
}
