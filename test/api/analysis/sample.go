package analysis

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Sample report.
var Sample = api.Analysis{
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
