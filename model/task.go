package model

import (
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
	"gorm.io/gorm"
	"time"
)

type TaskReport struct {
	Model
	Status    string
	Error     string
	Total     int
	Completed int
	Activity  JSON
	Result    JSON
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

type TaskGroup struct {
	Model
	BucketOwner
	Name  string
	Addon string
	Data  JSON
	Tasks []Task `gorm:"constraint:OnDelete:CASCADE"`
	List  JSON
	State string
}

//
// Propagate group data into the task.
func (m *TaskGroup) Propagate() (err error) {
	for i := range m.Tasks {
		task := &m.Tasks[i]
		task.State = m.State
		task.Bucket = m.Bucket
		if task.Addon == "" {
			task.Addon = m.Addon
		}
		if m.Data == nil {
			continue
		}
		a := Map{}
		err = json.Unmarshal(m.Data, &a)
		if err != nil {
			err = liberr.Wrap(
				err,
				"id",
				m.ID)
			return
		}
		b := Map{}
		err = json.Unmarshal(task.Data, &b)
		if err != nil {
			err = liberr.Wrap(
				err,
				"id",
				m.ID)
			return
		}
		task.Data, _ = json.Marshal(m.merge(a, b))
	}

	return
}

//
// merge maps B into A.
// The B map is the authority.
func (m *TaskGroup) merge(a, b Map) (out Map) {
	if a == nil {
		a = Map{}
	}
	if b == nil {
		b = Map{}
	}
	out = Map{}
	//
	// Merge-in elements found in B and in A.
	for k, v := range a {
		out[k] = v
		if bv, found := b[k]; found {
			out[k] = bv
			if av, cast := v.(Map); cast {
				if bv, cast := bv.(Map); cast {
					out[k] = m.merge(av, bv)
				} else {
					out[k] = bv
				}
			}
		}
	}
	//
	// Add elements found only in B.
	for k, v := range b {
		if _, found := a[k]; !found {
			out[k] = v
		}
	}

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
