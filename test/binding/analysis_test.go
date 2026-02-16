package binding

import (
	"sort"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"

	. "github.com/onsi/gomega"
)

func TestCreateGet(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application to be used for the test.
	app := &api.Application{Name: "Test"}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	analysis := api.Analysis{
		Commit: "abc123def456",
		Insights: []api.Insight{
			{
				RuleSet:     "eap7/eap8",
				Rule:        "javaee-to-jakartaee-00001",
				Name:        "Replace javax.* imports with jakarta.*",
				Description: "JavaEE has been rebranded to JakartaEE, requiring package name changes",
				Category:    "mandatory",
				Effort:      13,
				Labels:      []string{"konveyor.io/source=eap7", "konveyor.io/target=eap8"},
				Links: []api.Link{
					{
						URL:   "https://jakarta.ee/specifications/platform/9/",
						Title: "Jakarta EE 9 Platform Specification",
					},
				},
				Facts: api.Map{
					"package":  "javax.servlet",
					"severity": "high",
				},
				Incidents: []api.Incident{
					{
						File:     "src/main/java/com/example/MyServlet.java",
						Line:     5,
						Message:  "Replace 'import javax.servlet.*' with 'import jakarta.servlet.*'",
						CodeSnip: "import javax.servlet.http.HttpServlet;",
						Facts: api.Map{
							"kind": "import",
						},
					},
					{
						File:     "src/main/java/com/example/FilterConfig.java",
						Line:     8,
						Message:  "Replace 'import javax.servlet.Filter' with 'import jakarta.servlet.Filter'",
						CodeSnip: "import javax.servlet.Filter;",
						Facts: api.Map{
							"kind": "import",
						},
					},
				},
			},
			{
				RuleSet:     "quarkus/springboot",
				Rule:        "springboot-to-quarkus-00010",
				Name:        "Replace Spring Boot annotations with Quarkus equivalents",
				Description: "Spring Boot framework annotations need to be replaced with Quarkus/MicroProfile equivalents",
				Category:    "mandatory",
				Effort:      20,
				Labels:      []string{"konveyor.io/source=springboot", "konveyor.io/target=quarkus"},
				Links: []api.Link{
					{
						URL:   "https://quarkus.io/guides/spring-boot-properties",
						Title: "Quarkus Spring Boot Properties Guide",
					},
				},
				Facts: api.Map{
					"framework": "spring-boot",
				},
				Incidents: []api.Incident{
					{
						File:     "src/main/java/com/example/Application.java",
						Line:     12,
						Message:  "Replace '@SpringBootApplication' with '@QuarkusApplication'",
						CodeSnip: "@SpringBootApplication",
						Facts: api.Map{
							"annotation": "@SpringBootApplication",
						},
					},
					{
						File:     "src/main/java/com/example/RestController.java",
						Line:     15,
						Message:  "Replace '@RestController' with '@Path' and '@ApplicationScoped'",
						CodeSnip: "@RestController",
						Facts: api.Map{
							"annotation": "@RestController",
						},
					},
				},
			},
			{
				RuleSet:     "logging",
				Rule:        "logging-00001",
				Name:        "Use parameterized logging",
				Description: "String concatenation in logging statements should be replaced with parameterized logging",
				Category:    "potential",
				Effort:      12,
				Labels:      []string{"konveyor.io/target=cloud-readiness"},
				Links: []api.Link{
					{
						URL:   "https://www.slf4j.org/faq.html#logging_performance",
						Title: "SLF4J Performance FAQ",
					},
				},
				Facts: api.Map{
					"library": "slf4j",
				},
				Incidents: []api.Incident{
					{
						File:     "src/main/java/com/example/Service.java",
						Line:     42,
						Message:  "Use parameterized logging instead of string concatenation",
						CodeSnip: `log.info("Processing user: " + userName);`,
						Facts: api.Map{
							"pattern": "concatenation",
						},
					},
					{
						File:     "src/main/java/com/example/DataProcessor.java",
						Line:     78,
						Message:  "Use parameterized logging instead of string concatenation",
						CodeSnip: `log.debug("Record count: " + records.size());`,
						Facts: api.Map{
							"pattern": "concatenation",
						},
					},
				},
			},
		},
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "javax.servlet:servlet-api",
				Version:  "2.5",
				SHA:      "5959582d97d8b61f4d154ca9e495aafd16726e34",
				Indirect: false,
				Labels:   []string{"konveyor.io/dep=open-source"},
			},
			{
				Provider: "java",
				Name:     "org.springframework.boot:spring-boot-starter-web",
				Version:  "2.7.10",
				SHA:      "b8c9e82f7c8d9a1e5f3b2c4d8e7a9f0b1c2d3e4f",
				Indirect: false,
				Labels:   []string{"konveyor.io/dep=open-source", "konveyor.io/language=java"},
			},
			{
				Provider: "java",
				Name:     "commons-logging:commons-logging",
				Version:  "1.2",
				SHA:      "4bfc12adfe4842bf07b657f0369c4cb522955686",
				Indirect: true, // Transitive dependency
				Labels:   []string{"konveyor.io/dep=open-source"},
			},
		},
	}

	// CREATE: Create the analysis
	analysis.Application = api.Ref{ID: app.ID}
	err = client.Analysis.Create(&analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// LIST: List analyses and verify
	list, err := client.Analysis.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(BeNumerically(">", 0))
	found := false
	for _, a := range list {
		if a.ID == analysis.ID {
			found = true
			// Dependencies returned sorted by name
			sort.Slice(a.Dependencies, func(i, j int) bool {
				return a.Dependencies[i].Name < a.Dependencies[j].Name
			})
			sort.Slice(analysis.Dependencies, func(i, j int) bool {
				return analysis.Dependencies[i].Name < analysis.Dependencies[j].Name
			})
			eq, report := cmp.Eq(&analysis, &a,
				"Application.Name",
				"Insights.ID",
				"Insights.CreateTime",
				"Insights.Analysis",
				"Insights.Incidents.ID",
				"Insights.Incidents.CreateTime",
				"Insights.Incidents.Insight",
				"Dependencies.ID",
				"Dependencies.CreateTime",
				"Dependencies.Analysis",
			)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// GET: Retrieve the analysis and verify it matches
	retrieved, err := client.Analysis.Get(analysis.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(retrieved.Dependencies) == len(analysis.Dependencies)).To(BeTrue())
	// Dependencies returned sorted by name (already sorted from List section)
	sort.Slice(retrieved.Dependencies, func(i, j int) bool {
		return retrieved.Dependencies[i].Name < retrieved.Dependencies[j].Name
	})
	eq, report := cmp.Eq(
		analysis,
		retrieved,
		"Application.Name",
		"Insights.ID",
		"Insights.CreateTime",
		"Insights.Analysis",
		"Insights.Incidents.ID",
		"Insights.Incidents.CreateTime",
		"Insights.Incidents.Insight",
		"Dependencies.ID",
		"Dependencies.CreateTime",
		"Dependencies.Analysis",
	)
	g.Expect(eq).To(BeTrue(), report)
}

// TestAnalysisArchive tests archiving an analysis
func TestAnalysisArchive(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for the analysis
	app := &api.Application{Name: "Test App for Archive"}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis
	analysis := &api.Analysis{
		Application: api.Ref{ID: app.ID},
		Commit:      "archive-test-123",
		Effort:      50,
		Insights: []api.Insight{
			{
				RuleSet:     "test/archive",
				Rule:        "test-rule-001",
				Name:        "Test Insight",
				Description: "Test insight for archive",
				Category:    "mandatory",
				Effort:      10,
			},
		},
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "test-dependency",
				Version:  "1.0.0",
			},
		},
	}
	err = client.Analysis.Create(analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// ARCHIVE: Archive the analysis
	err = client.Analysis.Archive(analysis.ID)
	g.Expect(err).To(BeNil())

	// Note: Archive operation doesn't modify the analysis state in a way
	// we can verify via GET, but it should succeed without error
}

// TestAnalysisGlobalLists tests global list operations across all analyses
func TestAnalysisGlobalLists(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create first application
	app1 := &api.Application{Name: "Test App 1 for Global Lists"}
	err := client.Application.Create(app1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app1.ID)
	})

	// Create second application
	app2 := &api.Application{Name: "Test App 2 for Global Lists"}
	err = client.Application.Create(app2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app2.ID)
	})

	// CREATE: Create analysis for first app
	analysis1 := &api.Analysis{
		Application: api.Ref{ID: app1.ID},
		Commit:      "global-test-app1",
		Effort:      100,
		Insights: []api.Insight{
			{
				RuleSet:     "test/global",
				Rule:        "global-rule-001",
				Name:        "Global Test Insight 1",
				Description: "First insight for global testing",
				Category:    "mandatory",
				Effort:      20,
				Incidents: []api.Incident{
					{
						File:     "src/main/java/Test1.java",
						Line:     10,
						Message:  "Test incident 1",
						CodeSnip: "// test code 1",
					},
				},
			},
		},
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "global-dependency-1",
				Version:  "1.0.0",
			},
		},
	}
	err = client.Analysis.Create(analysis1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis1.ID)
	})

	// CREATE: Create analysis for second app
	analysis2 := &api.Analysis{
		Application: api.Ref{ID: app2.ID},
		Commit:      "global-test-app2",
		Effort:      150,
		Insights: []api.Insight{
			{
				RuleSet:     "test/global",
				Rule:        "global-rule-002",
				Name:        "Global Test Insight 2",
				Description: "Second insight for global testing",
				Category:    "potential",
				Effort:      30,
				Incidents: []api.Incident{
					{
						File:     "src/main/java/Test2.java",
						Line:     20,
						Message:  "Test incident 2",
						CodeSnip: "// test code 2",
					},
				},
			},
		},
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "global-dependency-2",
				Version:  "2.0.0",
			},
		},
	}
	err = client.Analysis.Create(analysis2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis2.ID)
	})

	// LIST DEPENDENCIES: Get global dependencies list
	deps, err := client.Analysis.ListDependencies()
	g.Expect(err).To(BeNil())
	g.Expect(len(deps)).To(BeNumerically(">", 0))
	// Verify our dependencies are in the list
	foundDep1 := false
	foundDep2 := false
	for _, dep := range deps {
		if dep.Name == "global-dependency-1" {
			foundDep1 = true
		}
		if dep.Name == "global-dependency-2" {
			foundDep2 = true
		}
	}
	g.Expect(foundDep1).To(BeTrue())
	g.Expect(foundDep2).To(BeTrue())

	// LIST INSIGHTS: Get global insights list
	insights, err := client.Analysis.ListInsights()
	g.Expect(err).To(BeNil())
	g.Expect(len(insights)).To(BeNumerically(">", 0))
	// Verify our insights are in the list
	foundInsight1 := false
	foundInsight2 := false
	var insight1ID uint
	for _, insight := range insights {
		if insight.Name == "Global Test Insight 1" {
			foundInsight1 = true
			insight1ID = insight.ID
		}
		if insight.Name == "Global Test Insight 2" {
			foundInsight2 = true
		}
	}
	g.Expect(foundInsight1).To(BeTrue())
	g.Expect(foundInsight2).To(BeTrue())

	// GET INSIGHT: Retrieve specific insight by ID
	if insight1ID > 0 {
		retrievedInsight, err := client.Analysis.GetInsight(insight1ID)
		g.Expect(err).To(BeNil())
		g.Expect(retrievedInsight).NotTo(BeNil())
		g.Expect(retrievedInsight.Name).To(Equal("Global Test Insight 1"))
	}

	// LIST INCIDENTS: Get global incidents list
	incidents, err := client.Analysis.ListIncidents()
	g.Expect(err).To(BeNil())
	g.Expect(len(incidents)).To(BeNumerically(">", 0))
	// Verify our incidents are in the list
	foundIncident1 := false
	foundIncident2 := false
	var incident1ID uint
	for _, incident := range incidents {
		if incident.Message == "Test incident 1" {
			foundIncident1 = true
			incident1ID = incident.ID
		}
		if incident.Message == "Test incident 2" {
			foundIncident2 = true
		}
	}
	g.Expect(foundIncident1).To(BeTrue())
	g.Expect(foundIncident2).To(BeTrue())

	// GET INCIDENT: Retrieve specific incident by ID
	if incident1ID > 0 {
		retrievedIncident, err := client.Analysis.GetIncident(incident1ID)
		g.Expect(err).To(BeNil())
		g.Expect(retrievedIncident).NotTo(BeNil())
		g.Expect(retrievedIncident.Message).To(Equal("Test incident 1"))
	}
}

// TestAnalysisSelect tests the Select() pattern for analysis-specific operations
func TestAnalysisSelect(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application
	app := &api.Application{Name: "Test App for Select"}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// CREATE: Create an analysis with multiple insights
	analysis := &api.Analysis{
		Application: api.Ref{ID: app.ID},
		Commit:      "select-test-123",
		Effort:      200,
		Insights: []api.Insight{
			{
				RuleSet:     "test/select",
				Rule:        "select-rule-001",
				Name:        "Select Test Insight 1",
				Description: "First insight for select testing",
				Category:    "mandatory",
				Effort:      40,
				Incidents: []api.Incident{
					{
						File:     "src/main/java/Select1.java",
						Line:     15,
						Message:  "Select incident 1A",
						CodeSnip: "// select code 1A",
					},
					{
						File:     "src/main/java/Select1.java",
						Line:     20,
						Message:  "Select incident 1B",
						CodeSnip: "// select code 1B",
					},
				},
			},
			{
				RuleSet:     "test/select",
				Rule:        "select-rule-002",
				Name:        "Select Test Insight 2",
				Description: "Second insight for select testing",
				Category:    "potential",
				Effort:      50,
				Incidents: []api.Incident{
					{
						File:     "src/main/java/Select2.java",
						Line:     25,
						Message:  "Select incident 2A",
						CodeSnip: "// select code 2A",
					},
				},
			},
		},
		Dependencies: []api.TechDependency{
			{
				Provider: "java",
				Name:     "select-dependency",
				Version:  "1.0.0",
			},
		},
	}
	err = client.Analysis.Create(analysis)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Analysis.Delete(analysis.ID)
	})

	// SELECT: Get selected analysis API
	selected := client.Analysis.Select(analysis.ID)

	// LIST INSIGHTS: Get insights for this specific analysis
	insights, err := selected.ListInsights()
	g.Expect(err).To(BeNil())
	g.Expect(len(insights)).To(Equal(2))

	// Verify insights belong to this analysis
	for _, insight := range insights {
		g.Expect(insight.Analysis).To(Equal(analysis.ID))
	}

	// Find the insight IDs
	var insight1ID uint
	for _, insight := range insights {
		if insight.Name == "Select Test Insight 1" {
			insight1ID = insight.ID
			break
		}
	}
	g.Expect(insight1ID).NotTo(BeZero())

	// GET INSIGHT: Get specific insight and its incidents
	selectedInsight := selected.GetInsight(insight1ID)

	// LIST INCIDENTS: Get incidents for this specific insight
	incidents, err := selectedInsight.ListIncidents()
	g.Expect(err).To(BeNil())
	g.Expect(len(incidents)).To(Equal(2))

	// Verify incidents belong to this insight
	for _, incident := range incidents {
		g.Expect(incident.Insight).To(Equal(insight1ID))
	}

	// Verify incident content
	foundIncident1A := false
	foundIncident1B := false
	for _, incident := range incidents {
		if incident.Message == "Select incident 1A" {
			foundIncident1A = true
			g.Expect(incident.Line).To(Equal(15))
		}
		if incident.Message == "Select incident 1B" {
			foundIncident1B = true
			g.Expect(incident.Line).To(Equal(20))
		}
	}
	g.Expect(foundIncident1A).To(BeTrue())
	g.Expect(foundIncident1B).To(BeTrue())
}
