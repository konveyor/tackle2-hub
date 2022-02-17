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
	TaskID    uint `gorm:"uniqueIndex"`
	Task      *Task
}

type Task struct {
	Model
	Name       string `gorm:"index"`
	Addon      string `gorm:"index"`
	Locator    string `gorm:"index"`
	Image      string
	Isolated   bool
	Data       JSON
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
