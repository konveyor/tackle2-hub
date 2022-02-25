package model

import (
	"time"
)

type TaskReport struct {
	Model
	Status    string
	Error     string
	Total     int
	Completed int
	Activity  JSON
	TaskID    uint `gorm:"<-:create;uniqueIndex"`
	Task      *Task
}

type Task struct {
	Model
	Name       string `gorm:"<-:create;index"`
	Addon      string `gorm:"<-:create;index"`
	Locator    string `gorm:"<-:create;index"`
	Image      string `gorm:"<-:create"`
	Isolated   bool   `gorm:"<-:create"`
	Data       JSON   `gorm:"<-:create"`
	Started    *time.Time
	Terminated *time.Time
	Status     string
	Error      string
	Job        string
	Report     *TaskReport `gorm:"constraint:OnDelete:CASCADE"`
}

func (m *Task) Reset() {
	m.Started = nil
	m.Terminated = nil
	m.Report = nil
	m.Status = ""
}
