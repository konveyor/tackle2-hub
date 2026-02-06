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

// TestReportAnalysisRuleReports tests the analysis rule reports
func TestReportAnalysisRuleReports(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create an application
	app := &api.Application{
		Name:        "Test App for Rule Reports",
		Description: "Application for testing rule reports",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with insights
	analysis := api.Analysis{
		Commit: "test-commit-123",
		Insights: []api.Insight{
			{
				RuleSet:     "eap7/eap8",
				Rule:        "test-rule-001",
				Name:        "Test Rule",
				Description: "Test rule description",
				Category:    "mandatory",
				Effort:      5,
				Labels:      []string{"konveyor.io/source=test"},
				Incidents: []api.Incident{
					{
						File:     "src/test/Test.java",
						Line:     10,
						Message:  "Test incident",
						CodeSnip: "test code",
					},
				},
			},
		},
	}
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// GET RULE REPORTS: Get the analysis rule reports
	ruleReports, err := client.Report.Analysis.RuleReports()
	g.Expect(err).To(BeNil())
	g.Expect(ruleReports).NotTo(BeNil())
	g.Expect(len(ruleReports)).To(BeNumerically(">", 0))
}

// TestReportAnalysisAppInsightReports tests the application insight reports
func TestReportAnalysisAppInsightReports(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create an application
	app := &api.Application{
		Name:        "Test App for Insight Reports",
		Description: "Application for testing insight reports",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with insights
	analysis := api.Analysis{
		Commit: "test-commit-456",
		Insights: []api.Insight{
			{
				RuleSet:     "quarkus/springboot",
				Rule:        "test-rule-002",
				Name:        "Test Spring Boot Rule",
				Description: "Test rule for Spring Boot migration",
				Category:    "mandatory",
				Effort:      8,
				Labels:      []string{"konveyor.io/source=springboot"},
				Incidents: []api.Incident{
					{
						File:     "src/main/java/App.java",
						Line:     5,
						Message:  "Test Spring Boot incident",
						CodeSnip: "test spring code",
					},
				},
			},
		},
	}
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// GET APP INSIGHT REPORTS: Get the insight reports for the application
	insightReports, err := client.Report.Analysis.AppInsightReports(app.ID)
	g.Expect(err).To(BeNil())
	g.Expect(insightReports).NotTo(BeNil())
	g.Expect(len(insightReports)).To(BeNumerically(">", 0))
}

// TestReportAnalysisInsightAppReports tests the insight application reports
func TestReportAnalysisInsightAppReports(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create an application
	app := &api.Application{
		Name:        "Test App for Insight App Reports",
		Description: "Application for testing insight app reports",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with insights
	analysis := api.Analysis{
		Commit: "test-commit-789",
		Insights: []api.Insight{
			{
				RuleSet:     "logging",
				Rule:        "test-rule-003",
				Name:        "Test Logging Rule",
				Description: "Test rule for logging improvements",
				Category:    "potential",
				Effort:      3,
				Labels:      []string{"konveyor.io/target=cloud-readiness"},
				Incidents: []api.Incident{
					{
						File:     "src/main/java/Logger.java",
						Line:     20,
						Message:  "Test logging incident",
						CodeSnip: "test logging code",
					},
				},
			},
		},
	}
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// GET INSIGHT APP REPORTS: Get the application reports for insights
	insightAppReports, err := client.Report.Analysis.InsightAppReports()
	g.Expect(err).To(BeNil())
	g.Expect(insightAppReports).NotTo(BeNil())
	g.Expect(len(insightAppReports)).To(BeNumerically(">", 0))
}

// TestReportAnalysisFileReports tests the insight file reports
func TestReportAnalysisFileReports(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create an application
	app := &api.Application{
		Name:        "Test App for File Reports",
		Description: "Application for testing file reports",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with insights and multiple incidents
	analysis := api.Analysis{
		Commit: "test-commit-abc",
		Insights: []api.Insight{
			{
				RuleSet:     "test-ruleset",
				Rule:        "test-rule-004",
				Name:        "Test File Rule",
				Description: "Test rule for file reports",
				Category:    "mandatory",
				Effort:      10,
				Labels:      []string{"konveyor.io/test"},
				Incidents: []api.Incident{
					{
						File:     "src/main/java/FileA.java",
						Line:     15,
						Message:  "Test incident in FileA",
						CodeSnip: "test code A",
					},
					{
						File:     "src/main/java/FileB.java",
						Line:     25,
						Message:  "Test incident in FileB",
						CodeSnip: "test code B",
					},
				},
			},
		},
	}
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// GET: Retrieve the analysis to get populated insight IDs
	retrieved, err := client.Analysis.Get(analysis.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(retrieved.Insights)).To(BeNumerically(">", 0))
	insightId := retrieved.Insights[0].ID

	// GET FILE REPORTS: Get the file reports for the insight
	fileReports, err := client.Report.Analysis.FileReports(insightId)
	g.Expect(err).To(BeNil())
	g.Expect(fileReports).NotTo(BeNil())
	g.Expect(len(fileReports)).To(BeNumerically(">", 0))
}

// TestReportAnalysisDepReports tests the dependency reports
func TestReportAnalysisDepReports(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create an application
	app := &api.Application{
		Name:        "Test App for Dep Reports",
		Description: "Application for testing dependency reports",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with dependencies
	analysis := api.Analysis{
		Commit: "test-commit-def",
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "org.springframework.boot:spring-boot-starter-web",
				Version:  "2.7.10",
				SHA:      "abc123def456",
				Indirect: false,
				Labels:   []string{"konveyor.io/dep=open-source"},
			},
			{
				Provider: "java",
				Name:     "commons-logging:commons-logging",
				Version:  "1.2",
				SHA:      "def456abc123",
				Indirect: true,
				Labels:   []string{"konveyor.io/dep=open-source"},
			},
		},
	}
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// GET DEP REPORTS: Get the dependency reports
	depReports, err := client.Report.Analysis.DepReports()
	g.Expect(err).To(BeNil())
	g.Expect(depReports).NotTo(BeNil())
	g.Expect(len(depReports)).To(BeNumerically(">", 0))
}

// TestReportAnalysisDepAppReports tests the dependency application reports
func TestReportAnalysisDepAppReports(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create an application
	app := &api.Application{
		Name:        "Test App for Dep App Reports",
		Description: "Application for testing dependency app reports",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with dependencies
	analysis := api.Analysis{
		Commit: "test-commit-ghi",
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "javax.servlet:servlet-api",
				Version:  "2.5",
				SHA:      "ghi789jkl012",
				Indirect: false,
				Labels:   []string{"konveyor.io/dep=open-source"},
			},
			{
				Provider: "java",
				Name:     "org.apache.logging.log4j:log4j-core",
				Version:  "2.17.1",
				SHA:      "jkl012mno345",
				Indirect: false,
				Labels:   []string{"konveyor.io/dep=open-source", "konveyor.io/language=java"},
			},
		},
	}
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// GET DEP APP REPORTS: Get the application reports for dependencies
	depAppReports, err := client.Report.Analysis.DepAppReports()
	g.Expect(err).To(BeNil())
	g.Expect(depAppReports).NotTo(BeNil())
	g.Expect(len(depAppReports)).To(BeNumerically(">", 0))
}
