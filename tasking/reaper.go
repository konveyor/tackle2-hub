package tasking

import (
	"context"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	pathlib "path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	ReaperUnit = time.Hour
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
func (r *TaskReaper) Run() {
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
		task := &list[i]
		deleted, err := r.delete(task)
		if deleted || err != nil {
			continue
		}
		err = r.emptyBucket(task)
		if err != nil {
			continue
		}
		err = r.deleteJob(task)
		if err != nil {
			continue
		}
	}
}

//
// delete Deletes the task as needed.
func (r *TaskReaper) delete(task *model.Task) (deleted bool, err error) {
	if !r.mayDelete(task) {
		return
	}
	result := r.DB.Delete(task)
	if result.Error == nil {
		Log.Info("Task deleted.", "id", task.ID)
	} else {
		Log.Trace(result.Error)
	}

	return
}

//
// emptyBucket Empties the task bucket as needed.
func (r *TaskReaper) emptyBucket(task *model.Task) (err error) {
	if !r.mayEmptyBucket(task) {
		return
	}
	err = task.EmptyBucket()
	if err == nil {
		Log.Info("Task bucket emptied.", "id", task.ID)
	} else {
		Log.Trace(err)
	}
	return
}

//
// deleteJob Deletes the associated job as needed.
func (r *TaskReaper) deleteJob(task *model.Task) (err error) {
	if !r.mayDeleteJob(task) {
		return
	}
	job := &batch.Job{}
	ns, name := pathlib.Split(task.Job)
	err = r.Client.Get(
		context.TODO(),
		client.ObjectKey{
			Namespace: ns,
			Name:      name,
		},
		job)
	if err != nil {
		if errors.IsNotFound(err) {
			err = nil
		} else {
			Log.Trace(err)
		}
		return
	}
	err = r.Client.Delete(context.TODO(), job)
	if err == nil {
		Log.Info("Task job deleted.", "id", task.ID)
	} else {
		Log.Trace(err)
	}
	return
}

//
// mayDelete determines if a task may be deleted.
// May be deleted when:
//   - Not associated with an application.
//   - Never submitted or terminated for defined period.
func (r *TaskReaper) mayDelete(task *model.Task) (approved bool) {
	if task.ApplicationID != nil {
		return
	}
	switch task.State {
	case Created:
		mark := task.CreateTime
		d := time.Duration(
			Settings.Hub.Task.Reaper.Created) * ReaperUnit
		approved = time.Since(mark) > d
	case Succeeded:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Succeeded) * ReaperUnit
		approved = time.Since(mark) > d
	case Failed:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Failed) * ReaperUnit
		approved = time.Since(mark) > d
	}
	return
}

//
// mayEmptyBucket Determines if a task bucket may be emptied.
// May be purged when:
//   - Not associated with a group.
//   - Terminated for defined period.
func (r *TaskReaper) mayEmptyBucket(task *model.Task) (may bool) {
	if task.TaskGroupID != nil {
		return
	}
	switch task.State {
	case Succeeded:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Succeeded) * ReaperUnit
		may = time.Since(mark) > d
	case Failed:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Failed) * ReaperUnit
		may = time.Since(mark) > d
	}
	return
}

//
// mayDeleteJob Determines if a task job can be deleted.
func (r *TaskReaper) mayDeleteJob(task *model.Task) (may bool) {
	switch task.State {
	case Succeeded:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Succeeded) * ReaperUnit
		may = time.Since(mark) > d
	}
	return
}

//
// GroupReaper reaps task groups.
type GroupReaper struct {
	// DB
	DB *gorm.DB
}

//
// Run Executes the reaper.
func (r *GroupReaper) Run() {
	list := []model.TaskGroup{}
	db := r.DB.Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		return
	}
	for i := range list {
		g := &list[i]
		deleted, err := r.delete(g)
		if deleted || err != nil {
			continue
		}
		err = r.emptyBucket(g)
		if err != nil {
			continue
		}
	}
}

//
// delete Deletes the group as needed.
func (r *GroupReaper) delete(g *model.TaskGroup) (deleted bool, err error) {
	if !r.mayDelete(g) {
		return
	}
	result := r.DB.Delete(g)
	if result.Error == nil {
		Log.Info("Group deleted.", "id", g.ID)
	} else {
		Log.Trace(result.Error)
	}
	return
}

//
// emptyBucket Empty the group bucket as needed.
func (r *GroupReaper) emptyBucket(g *model.TaskGroup) (err error) {
	if !r.mayDelete(g) {
		return
	}
	err = g.EmptyBucket()
	if err == nil {
		Log.Info("Group bucket emptied.", "id", g.ID)
	} else {
		Log.Trace(err)
	}
	return
}

//
// mayDelete Determines if a group may be deleted.
// May be deleted when:
//   - Empty for defined period.
func (r *GroupReaper) mayDelete(g *model.TaskGroup) (approved bool) {
	empty := len(g.Tasks) == 0
	mark := g.CreateTime
	d := time.Duration(
		Settings.Hub.Task.Reaper.Created) * ReaperUnit
	approved = empty && time.Since(mark) > d
	return
}

//
// mayEmptyBucket Determines if a group bucket may be emptied.
// May be purged when:
//   - All tasks buckets may be emptied.
func (r *GroupReaper) mayEmptyBucket(g *model.TaskGroup) (approved bool) {
	nMayPurge := 0
	tr := TaskReaper{DB: r.DB}
	for i := range g.Tasks {
		task := &g.Tasks[i]
		task.TaskGroupID = nil
		if tr.mayEmptyBucket(task) {
			nMayPurge++
		}
	}
	approved = nMayPurge == len(g.Tasks)
	return
}
