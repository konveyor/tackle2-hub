package model

import (
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
	Error         string
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
	if m.TaskGroupID == nil {
		err = m.BucketOwner.BeforeCreate(db)
	}
	m.Reset()
	return
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
