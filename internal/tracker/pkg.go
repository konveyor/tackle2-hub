package tracker

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
)

// Tracker types
const (
	JiraCloud  = "jira-cloud"
	JiraOnPrem = "jira-onprem"
)

// Ticket status
const (
	New        = "New"
	InProgress = "In Progress"
	Done       = "Done"
	Unknown    = "Unknown"
)

// Auth kinds
const (
	BearerAuth = "bearer"
	BasicAuth  = "basic-auth"
)

// Connector is a connector for an external ticket tracker.
type Connector interface {
	// With updates the connector with the tracker model.
	With(t *model.Tracker)
	// Create a ticket in the external tracker.
	Create(t *model.Ticket) error
	// RefreshAll refreshes the status of all tickets.
	RefreshAll() (map[*model.Ticket]bool, error)
	// TestConnection to the external ticket tracker.
	TestConnection() (bool, error)
	// Projects lists the tracker's projects.
	Projects() ([]Project, error)
	// Project gets a project from the tracker.
	Project(id string) (Project, error)
	// IssueTypes gets the issue types for a project.
	IssueTypes(id string) ([]IssueType, error)
}

// NewConnector instantiates a connector for an external ticket tracker.
func NewConnector(t *model.Tracker) (conn Connector, err error) {
	switch t.Kind {
	case JiraCloud, JiraOnPrem:
		conn = &JiraConnector{}
		conn.With(t)
	default:
		err = liberr.New("not implemented")
	}
	return
}

// Project represents an external ticket tracker's project
// in which an issue can be created.
type Project struct {
	ID   string
	Name string
}

// IssueType represents a type of issue that can be created on
// an external issue tracker.
type IssueType struct {
	ID   string
	Name string
}
