package reaper

import (
	"encoding/json"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

//
// TaskReaper reaps tasks.
type TaskReaper struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
}

//
// Run Executes the reaper.
// Rules by state:
//   Created
//   - Deleted after TTL.Created > created timestamp or
//     settings.Task.Reaper.Created.
//   Pending
//   - Deleted after TTL.Pending > created timestamp or
//     settings.Task.Reaper.Created.
//   Postponed
//   - Deleted after TTL.Postponed > created timestamp or
//     settings.Task.Reaper.Created.
//   Running
//   - Deleted after TTL.Running > started timestamp.
//   Succeeded
//   - Deleted after TTL > terminated timestamp or
//     settings.Task.Reaper.Succeeded.
//   - Bucket is released after the defined period.
//   - Pod is deleted after the defined period.
//   Failed
//   - Deleted after TTL > terminated timestamp or
//     settings.Task.Reaper.Failed.
//   - Bucket is released after the defined period.
//   - Pod is deleted after the defined period.
func (r *TaskReaper) Run() {
	Log.V(1).Info("Reaping tasks.")
	list := []model.Task{}
	result := r.DB.Find(
		&list,
		"state IN ?",
		[]string{
			task.Created,
			task.Succeeded,
			task.Failed,
		})
	Log.Error(result.Error, "")
	if result.Error != nil {
		return
	}
	for i := range list {
		m := &list[i]
		ttl := r.TTL(m)
		switch m.State {
		case task.Created:
			mark := m.CreateTime
			if ttl.Created > 0 {
				d := time.Duration(ttl.Created) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			} else {
				d := time.Duration(Settings.Hub.Task.Reaper.Created) * Unit
				if time.Since(mark) > d {
					r.release(m)
				}
			}
		case task.Pending:
			mark := m.CreateTime
			if ttl.Pending > 0 {
				d := time.Duration(ttl.Pending) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			}
		case task.Postponed:
			mark := m.CreateTime
			if ttl.Postponed > 0 {
				d := time.Duration(ttl.Postponed) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			}
		case task.Running:
			mark := *m.Started
			if ttl.Running > 0 {
				d := time.Duration(ttl.Running) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			}
		case task.Succeeded:
			mark := *m.Terminated
			if ttl.Succeeded > 0 {
				d := time.Duration(ttl.Succeeded) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			} else {
				d := time.Duration(Settings.Hub.Task.Reaper.Succeeded) * Unit
				if time.Since(mark) > d {
					r.release(m)
				}
			}
		case task.Failed:
			mark := *m.Terminated
			if ttl.Succeeded > 0 {
				d := time.Duration(ttl.Failed) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			} else {
				d := time.Duration(Settings.Hub.Task.Reaper.Failed) * Unit
				if time.Since(mark) > d {
					r.release(m)
				}
			}
		}
	}
}

//
// release resources.
func (r *TaskReaper) release(m *model.Task) {
	nChanged := 0
	if m.Pod != "" {
		rt := Task{Task: m}
		err := rt.Delete(r.Client)
		if err == nil {
			m.Pod = ""
			nChanged++
		} else {
			Log.Error(err, "")
		}
	}
	if m.HasBucket() {
		Log.Info("Task bucket released.", "id", m.ID)
		m.SetBucket(nil)
		nChanged++
	}
	if nChanged > 0 {
		err := r.DB.Save(m).Error
		if err != nil {
			Log.Error(err, "")
		}
	}
	return
}

//
// delete task.
func (r *TaskReaper) delete(m *model.Task) {
	rt := Task{m}
	err := rt.Delete(r.Client)
	if err != nil {
		Log.Error(err, "")
	}
	err = r.DB.Select(clause.Associations).Delete(m).Error
	if err == nil {
		Log.Info("Task deleted.", "id", m.ID)
	} else {
		Log.Error(err, "")
	}
}

//
// TTL returns the task TTL.
func (r *TaskReaper) TTL(m *model.Task) (ttl api.TTL) {
	if m.TTL != nil {
		_ = json.Unmarshal(m.TTL, &ttl)
	}

	return
}

//
//

//
// GroupReaper reaps task groups.
type GroupReaper struct {
	// DB
	DB *gorm.DB
}

//
// Run Executes the reaper.
// Rules by state:
//   Created
//   - Deleted after the defined period.
//   Ready (submitted)
//   - Deleted when all of its task have been deleted.
//   - Bucket is released immediately.
func (r *GroupReaper) Run() {
	Log.V(1).Info("Reaping groups.")
	list := []model.TaskGroup{}
	db := r.DB.Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		return
	}
	for i := range list {
		m := &list[i]
		switch m.State {
		case task.Created:
			mark := m.CreateTime
			d := time.Duration(
				Settings.Hub.Task.Reaper.Created) * Unit
			if time.Since(mark) > d {
				r.delete(m)
			}
		case task.Ready:
			if len(m.Tasks) == 0 {
				r.delete(m)
				continue
			}
			if m.HasBucket() {
				r.release(m)
			}
		}
	}
}

//
// release resources.
func (r *GroupReaper) release(m *model.TaskGroup) {
	m.SetBucket(nil)
	err := r.DB.Save(m).Error
	if err == nil {
		Log.Info("Group bucket released.", "id", m.ID)
	} else {
		Log.Error(err, "")
	}
}

//
// delete task.
func (r *GroupReaper) delete(m *model.TaskGroup) {
	err := r.DB.Delete(m).Error
	if err == nil {
		Log.Info("Group deleted.", "id", m.ID)
	} else {
		Log.Error(err, "")
	}
}
