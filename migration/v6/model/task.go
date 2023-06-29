package model

import (
	"encoding/json"
	"fmt"
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
	Errors        JSON
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
// Error appends an error.
func (m *Task) Error(severity, description string, x ...interface{}) {
	var list []TaskError
	description = fmt.Sprintf(description, x...)
	te := TaskError{Severity: severity, Description: description}
	_ = json.Unmarshal(m.Errors, &list)
	list = append(list, te)
	m.Errors, _ = json.Marshal(list)
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
// TaskError used in Task.Errors.
type TaskError struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type TaskReport struct {
	Model
	Status    string
	Errors    JSON
	Total     int
	Completed int
	Activity  JSON `gorm:"type:json"`
	Result    JSON `gorm:"type:json"`
	TaskID    uint `gorm:"<-:create;uniqueIndex"`
	Task      *Task
}
