package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/secret"
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

// With updates the value of the Setting with the json representation
// of the `value` parameter.
func (r *Setting) With(value any) (err error) {
	r.Value, err = json.Marshal(value)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// As unmarshalls the value of the Setting into the `ptr` parameter.
func (r *Setting) As(ptr any) (err error) {
	err = json.Unmarshal(r.Value, ptr)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

type Bucket struct {
	Model
	Path       string `gorm:"<-:create;uniqueIndex"`
	Expiration *time.Time
}

func (m *Bucket) BeforeCreate(db *gorm.DB) (err error) {
	if m.Path == "" {
		uid := uuid.New()
		m.Path = path.Join(
			Settings.Hub.Bucket.Path,
			uid.String())
		err = os.MkdirAll(m.Path, 0777)
		if err != nil {
			err = liberr.Wrap(
				err,
				"path",
				m.Path)
		}
	}
	return
}

type BucketOwner struct {
	BucketID *uint `gorm:"index" ref:"bucket"`
	Bucket   *Bucket
}

func (m *BucketOwner) BeforeCreate(db *gorm.DB) (err error) {
	if !m.HasBucket() {
		b := &Bucket{}
		err = db.Create(b).Error
		m.SetBucket(&b.ID)
	}
	return
}

func (m *BucketOwner) SetBucket(id *uint) {
	m.BucketID = id
	m.Bucket = nil
}

func (m *BucketOwner) HasBucket() (b bool) {
	return m.BucketID != nil
}

type File struct {
	Model
	Name       string
	Path       string `gorm:"<-:create;uniqueIndex"`
	Expiration *time.Time
}

func (m *File) BeforeCreate(db *gorm.DB) (err error) {
	uid := uuid.New()
	m.Path = path.Join(
		Settings.Hub.Bucket.Path,
		".file",
		uid.String())
	err = os.MkdirAll(path.Dir(m.Path), 0777)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			m.Path)
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
	m.Errors = nil
}

func (m *Task) BeforeCreate(db *gorm.DB) (err error) {
	err = m.BucketOwner.BeforeCreate(db)
	m.Reset()
	return
}

// Error appends an error.
func (m *Task) Error(severity, description string, x ...any) {
	var list []TaskError
	description = fmt.Sprintf(description, x...)
	te := TaskError{Severity: severity, Description: description}
	_ = json.Unmarshal(m.Errors, &list)
	list = append(list, te)
	m.Errors, _ = json.Marshal(list)
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

// Propagate group data into the task.
func (m *TaskGroup) Propagate() (err error) {
	for i := range m.Tasks {
		task := &m.Tasks[i]
		task.State = m.State
		task.SetBucket(m.BucketID)
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
	Name        string `gorm:"index;unique;not null"`
	Description string
	User        string
	Password    string
	Key         string
	Settings    string
	Proxies     []Proxy `gorm:"constraint:OnDelete:SET NULL"`
}

// Encrypt sensitive fields.
// The ref identity is used to determine when sensitive fields
// have changed and need to be (re)encrypted.
func (r *Identity) Encrypt(ref *Identity) (err error) {
	passphrase := Settings.Encryption.Passphrase
	aes := secret.AESCFB{}
	aes.Use(passphrase)
	if r.Password != ref.Password {
		if r.Password != "" {
			r.Password, err = aes.Encrypt(r.Password)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	if r.Key != ref.Key {
		if r.Key != "" {
			r.Key, err = aes.Encrypt(r.Key)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	if r.Settings != ref.Settings {
		if r.Settings != "" {
			r.Settings, err = aes.Encrypt(r.Settings)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}

// Decrypt sensitive fields.
func (r *Identity) Decrypt() (err error) {
	passphrase := Settings.Encryption.Passphrase
	aes := secret.AESCFB{}
	aes.Use(passphrase)
	if r.Password != "" {
		r.Password, err = aes.Decrypt(r.Password)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	if r.Key != "" {
		r.Key, err = aes.Decrypt(r.Key)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	if r.Settings != "" {
		r.Settings, err = aes.Decrypt(r.Settings)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}
