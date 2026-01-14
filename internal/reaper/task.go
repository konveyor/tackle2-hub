package reaper

import (
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// TaskReaper reaps tasks.
type TaskReaper struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
}

// Run Executes the reaper.
// Rules by state:
//
//	Created
//	- Deleted after TTL.Created > created timestamp or
//	  settings.Task.Reaper.Created.
//	Pending
//	- Deleted after TTL.Pending > created timestamp or
//	  settings.Task.Reaper.Created.
//	Postponed
//	- Deleted after TTL.Postponed > created timestamp or
//	  settings.Task.Reaper.Created.
//	Running
//	- Deleted after TTL.Running > started timestamp.
//	Succeeded
//	- Deleted after TTL > terminated timestamp or
//	  settings.Task.Reaper.Succeeded.
//	- Bucket is released after the defined period.
//	- Pod is deleted after the defined period.
//	Failed
//	- Deleted after TTL > terminated timestamp or
//	  settings.Task.Reaper.Failed.
//	- Bucket is released after the defined period.
//	- Pod is deleted after the defined period.
func (r *TaskReaper) Run() {
	Log.V(1).Info("Reaping tasks.")
	list := []task.Task{}
	result := r.DB.Find(
		&list,
		"state IN ? and reaped = 0",
		[]string{
			task.Created,
			task.Ready,
			task.Pending,
			task.QuotaBlocked,
			task.Running,
			task.Succeeded,
			task.Failed,
			task.Canceled,
		})
	Log.Error(result.Error, "")
	if result.Error != nil {
		return
	}
	pipelineSet := task.PipelineSet{}
	err := pipelineSet.Load(r.DB)
	if err != nil {
		Log.Error(err, "")
		return
	}
	for i := range list {
		m := &list[i]
		switch m.State {
		case task.Created:
			mark := m.CreateTime
			if m.TTL.Created > 0 {
				d := time.Duration(m.TTL.Created) * Unit
				if time.Since(mark) > d {
					if !pipelineSet.Contains(m) {
						r.delete(m)
					}
				}
			} else {
				d := time.Duration(Settings.Hub.Task.Reaper.Created) * Unit
				if time.Since(mark) > d {
					if !pipelineSet.Contains(m) {
						r.release(m)
					}
				}
			}
		case task.Ready,
			task.Pending,
			task.QuotaBlocked:
			mark := m.CreateTime
			if m.TTL.Pending > 0 {
				d := time.Duration(m.TTL.Pending) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			}
		case task.Running:
			mark := m.CreateTime
			if m.Terminated != nil {
				mark = *m.Started
			}
			if m.TTL.Running > 0 {
				d := time.Duration(m.TTL.Running) * Unit
				if time.Since(mark) > d {
					r.delete(m)
				}
			}
		case task.Succeeded:
			mark := m.CreateTime
			if m.Terminated != nil {
				mark = *m.Terminated
			}
			if m.TTL.Succeeded > 0 {
				d := time.Duration(m.TTL.Succeeded) * Unit
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
			mark := m.CreateTime
			if m.Terminated != nil {
				mark = *m.Terminated
			}
			if m.TTL.Failed > 0 {
				d := time.Duration(m.TTL.Failed) * Unit
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

// release bucket and file resources.
func (r *TaskReaper) release(m *task.Task) {
	nChanged := 0
	if !m.Reaped {
		m.Reaped = true
		nChanged++
	}
	if m.HasBucket() {
		Log.Info("Task bucket released.", "id", m.ID)
		m.SetBucket(nil)
		nChanged++
	}
	if len(m.Attached) > 0 {
		m.Attached = nil
		nChanged++
	}
	if nChanged > 0 {
		m.Event(task.Released)
		err := r.DB.Save(m).Error
		if err != nil {
			Log.Error(err, "")
		}
	}
	return
}

// podDelete deletes the task pod.
func (r *TaskReaper) podDelete(m *task.Task) {
	if m.Pod == "" {
		return
	}
	err := m.Delete(r.Client)
	if err != nil {
		Log.Error(err, "")
		return
	}
	err = r.DB.Save(m).Error
	if err != nil {
		Log.Error(err, "")
	}
}

// delete task.
func (r *TaskReaper) delete(m *task.Task) {
	err := m.Delete(r.Client)
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
//

// GroupReaper reaps task groups.
type GroupReaper struct {
	// DB
	DB *gorm.DB
}

// Run Executes the reaper.
// Rules by state:
//
//	Created
//	- Deleted after the defined period.
//	Ready (submitted)
//	- Deleted when all of its task have been deleted.
//	- Bucket is released immediately.
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
			mark := m.CreateTime
			if time.Since(mark) > time.Hour {
				r.release(m)
				if len(m.Tasks) == 0 {
					r.delete(m)
				} else {
					r.release(m)
				}
			}
		}
	}
}

// release resources.
func (r *GroupReaper) release(m *model.TaskGroup) {
	nChanged := 0
	if m.HasBucket() {
		Log.Info("Group bucket released.", "id", m.ID)
		m.SetBucket(nil)
		nChanged++
	}
	if len(m.List) > 0 {
		m.List = nil
		nChanged++
	}
	if nChanged > 0 {
		err := r.DB.Save(m).Error
		if err != nil {
			Log.Error(err, "")
		}
	}
}

// delete task.
func (r *GroupReaper) delete(m *model.TaskGroup) {
	err := r.DB.Delete(m).Error
	if err == nil {
		Log.Info("Group deleted.", "id", m.ID)
	} else {
		Log.Error(err, "")
	}
}
