package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Task struct {
	Model
	BucketOwner
	Name          string `gorm:"index"`
	Addon         string `gorm:"index"`
	Locator       string `gorm:"index"`
	Priority      int
	Image         string
	Variant       string
	Policy        string
	TTL           JSON
	Data          JSON
	Started       *time.Time
	Terminated    *time.Time
	State         string `gorm:"index"`
	Events        JSON
	Pod           string `gorm:"index"`
	Retries       int
	Canceled      bool
	Report        *TaskReport `gorm:"constraint:OnDelete:CASCADE"`
	ApplicationID *uint
	Application   *Application
	TaskGroupID   *uint `gorm:"<-:create"`
	TaskGroup     *TaskGroup
}

func (m *Task) Reset() {
	m.Started = nil
	m.Terminated = nil
	m.Report = nil
}

func (m *Task) BeforeCreate(db *gorm.DB) (err error) {
	err = m.BucketOwner.BeforeCreate(db)
	m.Reset()
	return
}

//
// Event appends an event.
func (m *Task) Event(kind, origin, description string) {
	var events []TaskEvent
	_ = json.Unmarshal(m.Events, &events)
	events = append(
		events,
		TaskEvent{
			Kind:        kind,
			Origin:      origin,
			Description: description,
		})
	m.Events, _ = json.Marshal(events)
}

//
// Map alias.
type Map = map[string]interface{}

//
// TTL time-to-live.
type TTL struct {
	Created   int `json:"created,omitempty"`
	Pending   int `json:"pending,omitempty"`
	Postponed int `json:"postponed,omitempty"`
	Running   int `json:"running,omitempty"`
	Succeeded int `json:"succeeded,omitempty"`
	Failed    int `json:"failed,omitempty"`
}

//
// TaskEvent used in Task.Errors.
type TaskEvent struct {
	Kind        string `json:"kind"`
	Origin      string `json:"origin,omitempty"`
	Description string `json:"description"`
}
