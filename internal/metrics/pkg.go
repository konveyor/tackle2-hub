package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	AssessmentsInitiated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "konveyor_assessments_initiated_total",
		Help: "The total number of initiated assessments",
	})
	TasksInitiated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "konveyor_tasks_initiated_total",
		Help: "The total number of initiated tasks",
	})
	Applications = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "konveyor_applications_inventoried",
		Help: "The current number of applications in inventory",
	})
	IssuesExported = promauto.NewCounter(prometheus.CounterOpts{
		Name: "konveyor_issues_exported_total",
		Help: "The total number of issues exported to external trackers",
	})
)
