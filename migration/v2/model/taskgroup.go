package model

import (
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
)

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
