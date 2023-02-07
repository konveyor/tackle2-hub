package model

import (
	"gorm.io/gorm"
)

//
// Seed the database with models.
func Seed(db *gorm.DB) {
	settings := []Setting{
		{Key: "git.insecure.enabled", Value: JSON("false")},
		{Key: "svn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.dependencies.update.forced", Value: JSON("false")},
	}
	_ = db.Create(settings)

	jobFunctions := []JobFunction{
		{Name: "Business Analyst"},
		{Name: "Business Service Owner / Manager"},
		{Name: "Consultant"},
		{Name: "DBA"},
		{Name: "Developer / Software Engineer"},
		{Name: "IT Operations"},
		{Name: "Program Manager"},
		{Name: "Project Manager"},
		{Name: "Service Owner"},
		{Name: "Solution Architect"},
		{Name: "System Administration"},
		{Name: "Test Analyst / Manager"},
	}
	_ = db.Create(jobFunctions)

	tagTypes := []TagType{
		{Name: "Application Type", Rank: 6, Color: "#ec7a08", Tags: []Tag{{Name: "COTS"}, {Name: "In house"}, {Name: "SaaS"}}},
		{Name: "Data Center", Rank: 5, Color: "#2b9af3", Tags: []Tag{{Name: "Boston (USA)"}, {Name: "London (UK)"}, {Name: "Paris (FR)"}, {Name: "Sydney (AU)"}}},
		{Name: "Database", Rank: 4, Color: "#6ec664", Tags: []Tag{{Name: "DB2"}, {Name: "MongoDB"}, {Name: "Oracle"}, {Name: "PostgreSQL"}, {Name: "SQL Server"}}},
		{Name: "Language", Rank: 1, Color: "#009596", Tags: []Tag{{Name: "C# ASP .Net"}, {Name: "C++"}, {Name: "COBOL"}, {Name: "Java"}, {Name: "Javascript"}, {Name: "Python"}}},
		{Name: "Operating System", Rank: 2, Color: "#a18fff", Tags: []Tag{{Name: "RHEL 8"}, {Name: "Windows Server 2016"}, {Name: "Z/OS"}}},
		{Name: "Runtime", Rank: 3, Color: "#7d1007", Tags: []Tag{{Name: "EAP"}, {Name: "JWS"}, {Name: "Quarkus"}, {Name: "Spring Boot"}, {Name: "Tomcat"}, {Name: "WebLogic"}, {Name: "WebSphere"}}},
	}
	_ = db.Create(tagTypes)

	proxies := []Proxy{
		{Kind: "http", Host: "", Port: 0},
		{Kind: "https", Host: "", Port: 0},
	}
	_ = db.Create(proxies)

	return
}
