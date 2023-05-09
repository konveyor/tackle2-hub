package tracker

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/tracker/jira"
)

// Tracker types
const (
	JiraCloud      = "jira-cloud"
	JiraServer     = "jira-server"
	JiraDataCenter = "jira-datacenter"
)

// Connector is a connector for an external ticket tracker.
type Connector interface {
	// With updates the connector with the tracker model.
	With(t *model.Tracker)
	// Create a ticket in the external tracker.
	Create(t *model.Ticket) error
	// RefreshAll refreshes the status of all tickets.
	RefreshAll() (map[*model.Ticket]bool, error)
	// GetMetadata from the tracker (ticket types, projects, etc)
	GetMetadata() (model.Metadata, error)
	// TestConnection to the external ticket tracker.
	TestConnection() (bool, error)
}

// NewConnector instantiates a connector for an external ticket tracker.
func NewConnector(t *model.Tracker) (conn Connector, err error) {
	switch t.Kind {
	case JiraCloud, JiraServer, JiraDataCenter:
		conn = &jira.Connector{}
		conn.With(t)
	default:
		err = liberr.New("not implemented")
	}
	return
}
