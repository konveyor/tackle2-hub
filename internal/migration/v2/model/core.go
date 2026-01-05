package model

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"gorm.io/gorm"
)

// Model Base model.
type Model struct {
	ID         uint      `gorm:"<-:create;primaryKey"`
	CreateTime time.Time `gorm:"<-:create;autoCreateTime"`
	CreateUser string    `gorm:"<-:create"`
	UpdateUser string
}

type Setting struct {
	Model
	Key   string `gorm:"<-:create;uniqueIndex"`
	Value JSON   `gorm:"type:json"`
}

type BucketOwner struct {
	Bucket string `gorm:"index"`
}

func (m *BucketOwner) BeforeCreate(db *gorm.DB) (err error) {
	uid := uuid.New()
	m.Bucket = path.Join(
		Settings.Hub.Bucket.Path,
		uid.String())
	err = os.MkdirAll(m.Bucket, 0777)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			m.Bucket)
	}
	return
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
	TTL           JSON `gorm:"type:json"`
	Data          JSON `gorm:"type:json"`
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

// Map alias.
type Map = map[string]any

// TTL time-to-live.
type TTL struct {
	Created   int `json:"created,omitempty"`
	Pending   int `json:"pending,omitempty"`
	Postponed int `json:"postponed,omitempty"`
	Running   int `json:"running,omitempty"`
	Succeeded int `json:"succeeded,omitempty"`
	Failed    int `json:"failed,omitempty"`
}

type TaskGroup struct {
	Model
	BucketOwner
	Name  string
	Addon string
	Data  JSON   `gorm:"type:json"`
	Tasks []Task `gorm:"constraint:OnDelete:CASCADE"`
	List  JSON   `gorm:"type:json"`
	State string
}

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

type TaskReport struct {
	Model
	Status    string
	Error     string
	Total     int
	Completed int
	Activity  JSON `gorm:"type:json"`
	Result    JSON `gorm:"type:json"`
	TaskID    uint `gorm:"<-:create;uniqueIndex"`
	Task      *Task
}

// Proxy configuration.
// kind = (http|https)
type Proxy struct {
	Model
	Enabled    bool
	Kind       string `gorm:"uniqueIndex"`
	Host       string `gorm:"not null"`
	Port       int
	Excluded   JSON  `gorm:"type:json"`
	IdentityID *uint `gorm:"index"`
	Identity   *Identity
}

// Identity represents and identity with a set of credentials.
type Identity struct {
	Model
	Kind        string `gorm:"not null"`
	Name        string `gorm:"not null"`
	Description string
	User        string
	Password    string
	Key         string
	Settings    string
	Proxies     []Proxy `gorm:"constraint:OnDelete:SET NULL"`
}
