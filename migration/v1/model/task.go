package model

import (
	"github.com/konveyor/tackle2-hub/model"
	"time"
)

type TaskReport struct {
	Model
	Status    string
	Error     string
	Total     int
	Completed int
	Activity  model.JSON
	Result    model.JSON
	TaskID    uint `gorm:"<-:create;uniqueIndex"`
	Task      *Task
}

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
	TTL           model.JSON
	Data          model.JSON
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

type TaskGroup struct {
	Model
	BucketOwner
	Name  string
	Addon string
	Data  model.JSON
	Tasks []Task `gorm:"constraint:OnDelete:CASCADE"`
	List  model.JSON
	State string
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
