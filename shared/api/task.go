package api

import (
	"time"
)

// TTL time-to-live.
type TTL struct {
	Created   int `json:"created,omitempty" yaml:",omitempty"`
	Pending   int `json:"pending,omitempty" yaml:",omitempty"`
	Running   int `json:"running,omitempty" yaml:",omitempty"`
	Succeeded int `json:"succeeded,omitempty" yaml:",omitempty"`
	Failed    int `json:"failed,omitempty" yaml:",omitempty"`
}

// TaskPolicy scheduling policies.
type TaskPolicy struct {
	Isolated       bool `json:"isolated,omitempty" yaml:",omitempty"`
	PreemptEnabled bool `json:"preemptEnabled,omitempty" yaml:"preemptEnabled,omitempty"`
	PreemptExempt  bool `json:"preemptExempt,omitempty" yaml:"preemptExempt,omitempty"`
}

// TaskError used in Task.Errors.
type TaskError struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// TaskEvent task event.
type TaskEvent struct {
	Kind   string    `json:"kind"`
	Count  int       `json:"count"`
	Reason string    `json:"reason,omitempty" yaml:",omitempty"`
	Last   time.Time `json:"last"`
}

// Attachment file attachment.
type Attachment struct {
	ID       uint   `json:"id" binding:"required"`
	Name     string `json:"name,omitempty" yaml:",omitempty"`
	Activity int    `json:"activity,omitempty" yaml:",omitempty"`
}

// Task REST resource.
type Task struct {
	Resource    `yaml:",inline"`
	Name        string       `json:"name,omitempty" yaml:",omitempty"`
	Kind        string       `json:"kind,omitempty" yaml:",omitempty"`
	Addon       string       `json:"addon,omitempty" yaml:",omitempty"`
	Extensions  []string     `json:"extensions,omitempty" yaml:",omitempty"`
	State       string       `json:"state,omitempty" yaml:",omitempty"`
	Locator     string       `json:"locator,omitempty" yaml:",omitempty"`
	Priority    int          `json:"priority,omitempty" yaml:",omitempty"`
	Policy      TaskPolicy   `json:"policy,omitempty" yaml:",omitempty"`
	TTL         TTL          `json:"ttl,omitempty" yaml:",omitempty"`
	Data        any          `json:"data,omitempty" yaml:",omitempty"`
	Application *Ref         `json:"application,omitempty" yaml:",omitempty"`
	Platform    *Ref         `json:"platform,omitempty" yaml:",omitempty"`
	Bucket      *Ref         `json:"bucket,omitempty" yaml:",omitempty"`
	Pod         string       `json:"pod,omitempty" yaml:",omitempty"`
	Retries     int          `json:"retries,omitempty" yaml:",omitempty"`
	Started     *time.Time   `json:"started,omitempty" yaml:",omitempty"`
	Terminated  *time.Time   `json:"terminated,omitempty" yaml:",omitempty"`
	Events      []TaskEvent  `json:"events,omitempty" yaml:",omitempty"`
	Errors      []TaskError  `json:"errors,omitempty" yaml:",omitempty"`
	Activity    []string     `json:"activity,omitempty" yaml:",omitempty"`
	Attached    []Attachment `json:"attached" yaml:",omitempty"`
}

// TaskReport REST resource.
type TaskReport struct {
	Resource  `yaml:",inline"`
	Status    string       `json:"status"`
	Errors    []TaskError  `json:"errors,omitempty" yaml:",omitempty"`
	Total     int          `json:"total,omitempty" yaml:",omitempty"`
	Completed int          `json:"completed,omitempty" yaml:",omitempty"`
	Activity  []string     `json:"activity,omitempty" yaml:",omitempty"`
	Attached  []Attachment `json:"attached,omitempty" yaml:",omitempty"`
	Result    any          `json:"result,omitempty" yaml:",omitempty"`
	TaskID    uint         `json:"task"`
}

// TaskQueue report.
type TaskQueue struct {
	Total        int `json:"total"`
	Ready        int `json:"ready"`
	Postponed    int `json:"postponed"`
	QuotaBlocked int `json:"quotaBlocked"`
	Pending      int `json:"pending"`
	Running      int `json:"running"`
}

// TaskDashboard report.
type TaskDashboard struct {
	Resource    `yaml:",inline"`
	Name        string     `json:"name,omitempty" yaml:",omitempty"`
	Kind        string     `json:"kind,omitempty" yaml:",omitempty"`
	Addon       string     `json:"addon,omitempty" yaml:",omitempty"`
	State       string     `json:"state,omitempty" yaml:",omitempty"`
	Locator     string     `json:"locator,omitempty" yaml:",omitempty"`
	Application *Ref       `json:"application,omitempty" yaml:",omitempty"`
	Platform    *Ref       `json:"platform,omitempty" yaml:",omitempty"`
	Started     *time.Time `json:"started,omitempty" yaml:",omitempty"`
	Terminated  *time.Time `json:"terminated,omitempty" yaml:",omitempty"`
	Errors      int        `json:"errors,omitempty" yaml:",omitempty"`
}

// TaskGroup REST resource.
type TaskGroup struct {
	Resource   `yaml:",inline"`
	Name       string     `json:"name"`
	Kind       string     `json:"kind,omitempty" yaml:",omitempty"`
	Addon      string     `json:"addon,omitempty" yaml:",omitempty"`
	Extensions []string   `json:"extensions,omitempty" yaml:",omitempty"`
	State      string     `json:"state"`
	Priority   int        `json:"priority,omitempty" yaml:",omitempty"`
	Policy     TaskPolicy `json:"policy,omitempty" yaml:",omitempty"`
	Data       any        `json:"data" swaggertype:"object" binding:"required"`
	Bucket     *Ref       `json:"bucket,omitempty"`
	Tasks      []Task     `json:"tasks"`
}
