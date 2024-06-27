package json

import (
	"time"

	"gopkg.in/yaml.v2"
)

// Ref represents a FK.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name,omitempty" yaml:",omitempty"`
}

// Map alias.
type Map = map[string]any

// Any alias.
type Any any

// Data json any field.
type Data struct {
	Any
}

// Merge merges the other into self.
// Both must be a map.
func (d *Data) Merge(other Data) (merged bool) {
	b, isMap := d.AsMap()
	if !isMap {
		return
	}
	a, isMap := other.AsMap()
	if !isMap {
		return
	}
	d.Any = d.merge(a, b)
	merged = true
	return
}

// Merge maps B into A.
// The B map takes precedence.
func (d *Data) merge(a, b map[any]any) (out map[any]any) {
	if a == nil {
		a = make(map[any]any)
	}
	if b == nil {
		b = make(map[any]any)
	}
	out = make(map[any]any)
	for k, v := range a {
		out[k] = v
		if bv, found := b[k]; found {
			out[k] = bv
			if av, cast := v.(map[any]any); cast {
				if bv, cast := bv.(map[any]any); cast {
					out[k] = d.merge(av, bv)
				} else {
					out[k] = bv
				}
			}
		}
	}
	for k, v := range b {
		if _, found := a[k]; !found {
			out[k] = v
		}
	}

	return
}

// AsMap returns self as a map.
func (d *Data) AsMap() (mp map[any]any, isMap bool) {
	if d.Any == nil {
		return
	}
	b, err := yaml.Marshal(d.Any)
	if err != nil {
		return
	}
	mp = make(map[any]any)
	err = yaml.Unmarshal(b, &mp)
	if err != nil {
		return
	}
	isMap = true
	return
}

// Repository represents an SCM repository.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

// Link URL link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// ArchivedIssue resource created when issues are archived.
type ArchivedIssue struct {
	RuleSet     string `json:"ruleSet"`
	Rule        string `json:"rule"`
	Name        string `json:"name,omitempty" yaml:",omitempty"`
	Description string `json:"description,omitempty" yaml:",omitempty"`
	Category    string `json:"category"`
	Effort      int    `json:"effort"`
	Incidents   int    `json:"incidents"`
}

// TaskEvent task event.
type TaskEvent struct {
	Kind   string    `json:"kind"`
	Count  int       `json:"count"`
	Reason string    `json:"reason,omitempty" yaml:",omitempty"`
	Last   time.Time `json:"last"`
}

// TaskError used in Task.Errors.
type TaskError struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// TaskPolicy scheduling policy.
type TaskPolicy struct {
	Isolated       bool `json:"isolated,omitempty" yaml:",omitempty"`
	PreemptEnabled bool `json:"preemptEnabled,omitempty" yaml:"preemptEnabled,omitempty"`
	PreemptExempt  bool `json:"preemptExempt,omitempty" yaml:"preemptExempt,omitempty"`
}

// TTL time-to-live.
type TTL struct {
	Created   int `json:"created,omitempty" yaml:",omitempty"`
	Pending   int `json:"pending,omitempty" yaml:",omitempty"`
	Running   int `json:"running,omitempty" yaml:",omitempty"`
	Succeeded int `json:"succeeded,omitempty" yaml:",omitempty"`
	Failed    int `json:"failed,omitempty" yaml:",omitempty"`
}

// Attachment file attachment.
type Attachment struct {
	ID       uint   `json:"id" binding:"required"`
	Name     string `json:"name,omitempty" yaml:",omitempty"`
	Activity int    `json:"activity,omitempty" yaml:",omitempty"`
}

type TargetLabel struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}
