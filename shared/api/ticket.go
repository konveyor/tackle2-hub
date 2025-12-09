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
