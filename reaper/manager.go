package reaper

import (
	"context"
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"path"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const (
	Unit = time.Minute
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("reaper")
)

type Task = task.Task

//
// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
}

//
// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	registered := []Reaper{
		&TaskReaper{
			Client: m.Client,
			DB:     m.DB,
		},
		&GroupReaper{
			DB: m.DB,
		},
		&BucketReaper{
			DB: m.DB,
		},
		&FileReaper{
			DB: m.DB,
		},
	}
	go func() {
		Log.Info("Started.")
		defer Log.Info("Died.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, r := range registered {
					r.Run()
				}
				m.pause()
			}
		}
	}()
}

//
// Pause.
func (m *Manager) pause() {
	d := Unit * time.Duration(Settings.Frequency.Reaper)
	time.Sleep(d)
}

//
// Reaper interface.
type Reaper interface {
	Run()
}

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
	Log.Info("Reaping tasks.")
	list := []model.Task{}
	result := r.DB.Find(
		&list,
		"state IN ?",
		[]string{
			task.Created,
			task.Succeeded,
			task.Failed,
		})
	Log.Trace(result.Error)
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
			Log.Trace(err)
		}
	}
	if m.Bucket != "" {
		Log.Info("Task bucket released.", "id", m.ID)
		m.Bucket = ""
		nChanged++
	}
	if nChanged > 0 {
		err := r.DB.Save(m).Error
		if err != nil {
			Log.Trace(err)
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
		Log.Trace(err)
	}
	err = r.DB.Select(clause.Associations).Delete(m).Error
	if err == nil {
		Log.Info("Task deleted.", "id", m.ID)
	} else {
		Log.Trace(err)
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
			if m.Bucket != "" {
				r.release(m)
			}
		}
	}
}

//
// release resources.
func (r *GroupReaper) release(m *model.TaskGroup) {
	m.Bucket = ""
	err := r.DB.Save(m).Error
	if err == nil {
		Log.Info("Group bucket released.", "id", m.ID)
	} else {
		Log.Trace(err)
	}
}

//
// delete task.
func (r *GroupReaper) delete(m *model.TaskGroup) {
	err := r.DB.Delete(m).Error
	if err == nil {
		Log.Info("Group deleted.", "id", m.ID)
	} else {
		Log.Trace(err)
	}
}

//
// BucketReaper bucket reaper.
type BucketReaper struct {
	// DB
	DB *gorm.DB
}

//
// Run Executes the reaper.
// (.) dot prefixed directories are ignored.
// A bucket is deleted when it is no longer referenced.
func (r *BucketReaper) Run() {
	Log.V(1).Info("Reaping buckets.")
	entries, err := os.ReadDir(Settings.Hub.Bucket.Path)
	if err != nil {
		Log.Trace(err)
	}
	for _, dir := range entries {
		name := dir.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		bucket := path.Join(
			Settings.Hub.Bucket.Path,
			name)
		busy, err := r.busy(bucket)
		if err != nil {
			Log.Trace(err)
			continue
		}
		if busy {
			continue
		}
		err = r.delete(bucket)
		if err != nil {
			Log.Trace(err)
			continue
		}
	}
}

//
// busy determines if anything references the bucket.
func (r *BucketReaper) busy(bucket string) (busy bool, err error) {
	nRef := int64(0)
	var n int64
	for _, m := range []interface{}{
		&model.Application{},
		&model.TaskGroup{},
		&model.Task{},
	} {
		db := r.DB.Model(m)
		db = db.Where("Bucket", bucket)
		err = db.Count(&n).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		nRef += n
	}
	busy = nRef > 0
	return
}

//
// Delete bucket.
func (r *BucketReaper) delete(path string) (err error) {
	err = nas.RmDir(path)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	} else {
		Log.Info("Bucket deleted.", "path", path)
	}
	return
}
