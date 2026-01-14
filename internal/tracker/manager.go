package tracker

import (
	"context"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	Log = logr.New("tickets", 0)
)

// Intervals
const (
	IntervalCreateRetry  = time.Second * 30
	IntervalRefresh      = time.Second * 30
	IntervalConnected    = time.Second * 60
	IntervalDisconnected = time.Second * 10
)

// Manager provides ticket management.
type Manager struct {
	// DB
	DB *gorm.DB
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	go func() {
		Log.Info("Started.")
		defer Log.Info("Died.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				m.testConnections()
				m.refreshTickets()
				m.createPending()
			}
		}
	}()
}

// testConnections to external trackers.
func (m *Manager) testConnections() {
	var list []model.Tracker
	result := m.DB.Preload(clause.Associations).Find(&list)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query trackers.")
		return
	}
	for i := range list {
		tracker := &list[i]
		var ago time.Time
		if tracker.Connected {
			ago = tracker.LastUpdated.Add(IntervalConnected)
		} else {
			ago = tracker.LastUpdated.Add(IntervalDisconnected)
		}
		if ago.Before(time.Now()) {
			err := m.testConnection(tracker)
			if err != nil {
				Log.Error(err, "Failed to update tracker", "tracker", tracker.ID)
			}
		}
	}

	return
}

// testConnection to the external tracker.
func (m *Manager) testConnection(tracker *model.Tracker) (err error) {
	conn, err := NewConnector(tracker)
	if err != nil {
		return
	}
	connected, err := conn.TestConnection()
	if err != nil {
		Log.Error(err, "Connection test failed.", "tracker", tracker.ID)
		tracker.Message = err.Error()
		err = nil
	}

	if connected {
		tracker.Message = ""
	}
	tracker.Connected = connected
	tracker.LastUpdated = time.Now()

	result := m.DB.Save(tracker)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (m *Manager) refreshTickets() {
	var list []model.Tracker
	result := m.DB.Preload(clause.Associations).Where("connected = ?", true).Find(&list)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query trackers.")
		return
	}
	for i := range list {
		tracker := &list[i]
		ago := tracker.LastUpdated.Add(IntervalRefresh)
		if ago.Before(time.Now()) {
			err := m.refresh(tracker)
			if err != nil {
				Log.Error(err, "Failed to refresh tracker.", "tracker", tracker.ID)
			}
		}
	}
}

// Update the hub's representation of the ticket with fresh
// status information from the external tracker.
func (m *Manager) refresh(tracker *model.Tracker) (err error) {
	conn, err := NewConnector(tracker)
	if err != nil {
		return
	}
	tickets, err := conn.RefreshAll()
	if err != nil {
		return
	}
	for t, found := range tickets {
		if found {
			result := m.DB.Save(t)
			if result.Error != nil {
				Log.Error(result.Error, "Failed to save ticket.", "ticket", t.ID)
				continue
			}
		} else {
			result := m.DB.Delete(t)
			if result.Error != nil {
				Log.Error(result.Error, "Failed to delete ticket.", "ticket", t.ID)
				continue
			}
		}
	}

	return
}

// Create pending tickets.
func (m *Manager) createPending() {
	var list []model.Tracker
	result := m.DB.Preload(clause.Associations).Preload("Tickets.Application").Where("connected = ?", true).Find(&list)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query trackers.")
		return
	}
	for i := range list {
		tracker := &list[i]
		conn, err := NewConnector(tracker)
		if err != nil {
			Log.Error(err, "Unable to build connector for tracker.", "tracker", tracker.ID)
			continue
		}
		for j := range tracker.Tickets {
			t := &tracker.Tickets[j]
			ago := t.LastUpdated.Add(IntervalCreateRetry)
			// if the ticket has already been created, or if there was previously an error
			// creating it and the retry window has not yet passed, skip this ticket.
			if t.Created || (t.Error && !ago.Before(time.Now())) {
				continue
			}
			err = m.create(conn, t)
			if err != nil {
				Log.Error(err, "Failed to create ticket.", "ticket", t.ID)
			}
		}
	}
}

// Create the ticket in its tracker.
func (m *Manager) create(conn Connector, ticket *model.Ticket) (err error) {
	err = conn.Create(ticket)
	if err != nil {
		return
	}
	result := m.DB.Save(ticket)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}
