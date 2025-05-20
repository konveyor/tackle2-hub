package reaper

import (
	"time"

	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
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
	Log.Info("TaskReaper: beginning.")
	mark := time.Now()
	m := &task.Task{}
	db := r.DB.Model(m)
	db = db.Where(
		"state IN ?",
		[]string{
			task.Created,
			task.Succeeded,
			task.Failed,
			task.Canceled,
		})
	cursor, err := db.Rows()
	Log.Error(err, "")
	if err != nil {
		return
	}
	defer func() {
		_ = cursor.Close()
	}()
	pipelines, err := r.pipelineMap()
	if err != nil {
		Log.Error(err, "")
		return
	}
	for cursor.Next() {
		err = db.ScanRows(cursor, m)
		if err != nil {
			return
		}
		switch m.State {
		case task.Created:
			mark := m.CreateTime
			if m.TTL.Created > 0 {
				d := time.Duration(m.TTL.Created) * Unit
				if time.Since(mark) > d {
					if !r.inPipelined(pipelines, m) {
						r.delete(m)
					}
				}
			} else {
				d := time.Duration(Settings.Hub.Task.Reaper.Created) * Unit
				if time.Since(mark) > d {
					if !r.inPipelined(pipelines, m) {
						r.release(m)
					}
				}
			}
		case task.Pending:
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
			d := time.Duration(Settings.Hub.Task.Pod.Retention.Succeeded) * Unit
			if time.Since(mark) > d {
				r.podDelete(m)
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
			d := time.Duration(Settings.Hub.Task.Pod.Retention.Failed) * Unit
			if time.Since(mark) > d {
				r.podDelete(m)
			}
		}
	}

	Log.Info("TaskReaper: ended.", "duration", time.Since(mark))
}

// release bucket and file resources.
func (r *TaskReaper) release(m *task.Task) {
	nChanged := 0
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
	err = r.DB.Delete(m).Error
	if err == nil {
		Log.Info("Task deleted.", "id", m.ID)
	} else {
		Log.Error(err, "")
	}
}

// pipelines returns a map of TaskGroup.Mode keyed by ID.
func (r *TaskReaper) pipelineMap() (mp PipelineMap, err error) {
	var list []*model.TaskGroup
	mp = make(map[uint]string)
	db := r.DB.Select("ID", "Mode")
	err = db.Find(&list).Error
	if err != nil {
		return
	}
	for _, m := range list {
		if m.Mode == task.Pipeline {
			mp[m.ID] = m.Mode
		}
	}
	return
}

// inPipelined returns true when the task is part of a pipeline.
func (r *TaskReaper) inPipelined(mp PipelineMap, m *task.Task) (b bool) {
	if m.TaskGroupID != nil {
		_, b = mp[*m.TaskGroupID]
	}
	return
}

type PipelineMap = map[uint]string

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
	Log.Info("GroupReaper: beginning.")
	mark := time.Now()
	type M struct {
		*model.TaskGroup
		Count int64
	}
	m := &M{}
	db := r.DB.Table("TaskGroup g")
	db = db.Joins("LEFT JOIN Task t ON t.TaskGroupID = g.ID")
	db = db.Select(
		"g.*",
		"COUNT(t.id) Count")
	db = db.Group("g.ID")
	cursor, err := db.Rows()
	if err != nil {
		Log.Error(err, "")
		return
	}
	defer func() {
		_ = cursor.Close()
	}()
	for cursor.Next() {
		err = r.DB.ScanRows(cursor, m)
		if err != nil {
			return
		}
		switch m.State {
		case "", task.Created:
			mark := m.CreateTime
			d := time.Duration(
				Settings.Hub.Task.Reaper.Created) * Unit
			if time.Since(mark) > d {
				r.delete(m.TaskGroup)
			}
		case task.Ready:
			mark := m.CreateTime
			if time.Since(mark) > time.Hour {
				r.release(m.TaskGroup)
				if m.Count == 0 {
					r.delete(m.TaskGroup)
				} else {
					r.release(m.TaskGroup)
				}
			}
		}
	}

	Log.Info("GroupReaper: ended.", "duration", time.Since(mark))
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
