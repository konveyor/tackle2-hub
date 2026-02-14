package report

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) (h Report) {
	h.Analysis = Analysis{client: client}
	h.Task = Task{client: client}
	return
}

// Report API.
type Report struct {
	Analysis Analysis
	Task     Task
}

// Analysis report API.
type Analysis struct {
	client client.RestClient
}

// RuleReports returns analysis rule reports.
func (h Analysis) RuleReports() (list []api.RuleReport, err error) {
	list = []api.RuleReport{}
	err = h.client.Get(api.AnalysisReportRuleRoute, &list)
	return
}

// AppInsightReports returns insight reports for an application.
func (h Analysis) AppInsightReports(appId uint) (list []api.InsightReport, err error) {
	list = []api.InsightReport{}
	path := client.Path(api.AnalysisReportAppsInsightsRoute).Inject(client.Params{api.ID: appId})
	err = h.client.Get(path, &list)
	return
}

// InsightAppReports returns application reports for insights.
func (h Analysis) InsightAppReports() (list []api.InsightAppReport, err error) {
	list = []api.InsightAppReport{}
	err = h.client.Get(api.AnalysisReportInsightsAppsRoute, &list)
	return
}

// FileReports returns file reports for an insight.
func (h Analysis) FileReports(insightId uint) (list []api.FileReport, err error) {
	list = []api.FileReport{}
	path := client.Path(api.AnalysisReportFileRoute).Inject(client.Params{api.ID: insightId})
	err = h.client.Get(path, &list)
	return
}

// DepReports returns dependency reports.
func (h Analysis) DepReports() (list []api.DepReport, err error) {
	list = []api.DepReport{}
	err = h.client.Get(api.AnalysisReportDepsRoute, &list)
	return
}

// DepAppReports returns application reports for dependencies.
func (h Analysis) DepAppReports() (list []api.DepAppReport, err error) {
	list = []api.DepAppReport{}
	err = h.client.Get(api.AnalysisReportDepsAppsRoute, &list)
	return
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
