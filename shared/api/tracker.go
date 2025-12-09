package api

import (
	"time"
)

// Tracker API Resource
type Tracker struct {
	Resource    `yaml:",inline"`
	Name        string    `json:"name" binding:"required"`
	URL         string    `json:"url" binding:"required"`
	Kind        string    `json:"kind" binding:"required,oneof=jira-cloud jira-onprem"`
	Message     string    `json:"message"`
	Connected   bool      `json:"connected"`
	LastUpdated time.Time `json:"lastUpdated" yaml:"lastUpdated"`
	Identity    Ref       `json:"identity" binding:"required"`
	Insecure    bool      `json:"insecure"`
}

// Project API Resource
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IssueType API Resource
type IssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
