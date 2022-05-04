package tasking

import (
	"context"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	core "k8s.io/api/core/v1"
	"os"
	"os/exec"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const (
	ReaperUnit = time.Minute
)

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
	Client client.Client
}

//
// Run Executes the reaper.
// Rules by state:
//   Created
//   - Deleted after the defined period.
//   Succeeded
//   - Deleted after the defined period when not associated with
//     an application. Tasks associated with an application
//     are deleted when the application is deleted.
//   - Bucket is released after the defined period.
//   - Pod is deleted after the defined period.
//   Failed
//   - Deleted after the defined period when not associated with
//     an application. Tasks associated with an application
//     are deleted when the application is deleted.
//   - Bucket is released after the defined period.
//   - Pod is deleted after the defined period.
func (r *TaskReaper) Run() {
	Log.V(1).Info("Reaping tasks.")
	list := []model.Task{}
	result := r.DB.Find(
		&list,
		"state IN ?",
		[]string{
			Created,
			Succeeded,
			Failed,
		})
	Log.Trace(result.Error)
	if result.Error != nil {
		return
	}
	for i := range list {
		m := &list[i]
		switch m.State {
		case Created:
			mark := m.CreateTime
			d := time.Duration(
				Settings.Hub.Task.Reaper.Created) * ReaperUnit
			if time.Since(mark) > d {
				r.delete(m)
			}
		case Succeeded:
			mark := *m.Terminated
			d := time.Duration(
				Settings.Hub.Task.Reaper.Succeeded) * ReaperUnit
			if time.Since(mark) > d {
				if m.ApplicationID == nil {
					r.delete(m)
				} else {
					r.release(m)
				}
			}
		case Failed:
			mark := *m.Terminated
			d := time.Duration(
				Settings.Hub.Task.Reaper.Failed) * ReaperUnit
			if time.Since(mark) > d {
				if m.ApplicationID == nil {
					r.delete(m)
				} else {
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
		err := r.deletePod(m)
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
// deletePod Deletes the associated pod as needed.
func (r *TaskReaper) deletePod(m *model.Task) (err error) {
	if m.Pod == "" {
		return
	}
	pod := &core.Pod{}
	pod.Namespace = path.Dir(m.Pod)
	pod.Name = path.Base(m.Pod)
	err = r.Client.Delete(context.TODO(), pod)
	if err == nil {
		Log.Info(
			"Task pod deleted.",
			"id",
			m.ID,
			"pod",
			pod.Name)
	} else {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// delete task.
func (r *TaskReaper) delete(m *model.Task) {
	err := r.deletePod(m)
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
		case Created:
			mark := m.CreateTime
			d := time.Duration(
				Settings.Hub.Task.Reaper.Created) * ReaperUnit
			if time.Since(mark) > d {
				r.delete(m)
			}
		case Ready:
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
	cmd := exec.Command("/usr/bin/rm", "-rf", path)
	b, err := cmd.CombinedOutput()
	if err != nil {
		err = liberr.New(
			string(b),
			"path",
			path)
	} else {
		Log.Info("Bucket deleted.", "path", path)
	}
	return
}
