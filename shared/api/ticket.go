package api

import (
	"time"
)

// Ticket API Resource
type Ticket struct {
	Resource    `yaml:",inline"`
	Kind        string    `json:"kind" binding:"required"`
	Reference   string    `json:"reference"`
	Link        string    `json:"link"`
	Parent      string    `json:"parent" binding:"required"`
	Error       bool      `json:"error"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated" yaml:"lastUpdated"`
	Fields      Map       `json:"fields"`
	Application Ref       `json:"application" binding:"required"`
	Tracker     Ref       `json:"tracker" binding:"required"`
}

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
