package binding

import (
	"testing"

	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

// TestReportTaskQueued tests the task queue report
func TestReportTaskQueued(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Skip("NEEDS CLUSTER SIMULATOR") // TODO: add hub k8s simulator.
	return

	// CREATE: Create a few tasks in different states
	task1 := &api.Task{
		Name:  "Test Task 1 for Queue Report",
		Addon: "analyzer",
		Kind:  "test-kind",
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
		State:    tasking.Ready,
		Priority: 5,
	}
	err := client.Task.Create(task1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Task.Delete(task1.ID)
	})

	task2 := &api.Task{
		Name:  "Test Task 2 for Queue Report",
		Addon: "analyzer",
		Kind:  "test-kind",
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
		State:    tasking.Pending,
		Priority: 3,
	}
	err = client.Task.Create(task2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Task.Delete(task2.ID)
	})

	// GET QUEUED REPORT: Get the task queue report
	queuedReport, err := client.Report.Task.Queued()
	g.Expect(err).To(BeNil())
	g.Expect(queuedReport).NotTo(BeNil())
	g.Expect(queuedReport.Total).To(BeNumerically(">=", 2))
}

// TestReportTaskDashboard tests the task dashboard report
func TestReportTaskDashboard(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Skip("NEEDS CLUSTER SIMULATOR") // TODO: add hub k8s simulator.
	return

	// CREATE: Create a task for the dashboard report
	task := &api.Task{
		Name:  "Test Task for Dashboard Report",
		Addon: "analyzer",
		Kind:  "test-kind",
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
		State:    tasking.Created,
		Priority: 5,
	}
	err := client.Task.Create(task)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Task.Delete(task.ID)
	})

	// GET DASHBOARD REPORT: Get the task dashboard report
	dashboardReport, err := client.Report.Task.Dashboard()
	g.Expect(err).To(BeNil())
	g.Expect(dashboardReport).NotTo(BeNil())
	g.Expect(len(dashboardReport)).To(BeNumerically(">=", 1))

	// Verify at least one task in the dashboard matches our created task
	found := false
	for _, dashTask := range dashboardReport {
		if dashTask.ID == task.ID {
			found = true
			g.Expect(dashTask.Name).To(Equal(task.Name))
			g.Expect(dashTask.Addon).To(Equal(task.Addon))
			g.Expect(dashTask.Kind).To(Equal(task.Kind))
			break
		}
	}
	g.Expect(found).To(BeTrue())
}
